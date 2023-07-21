// Package messaging contains repository implementations for exchanging events in an enterprise environment
package messaging

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/Azure/go-amqp"
	"github.com/golang/glog"
)

const (
	// UMBUserEventsTopic is the name of the topic that publishes user events from the UMB
	UMBUserEventsTopic string = "VirtualTopic.canonical.user"
)

// UMBMessageBusRepository can send and receive events on the Unified Message Bus
type UMBMessageBusRepository struct {
	config      serviceconfig.UMBConfig
	conn        *amqp.Conn
	subjectRecv *amqp.Receiver
	recvCtx     context.Context
	recvCancel  context.CancelFunc
	errs        chan error
	changes     chan contracts.SubjectAddOrUpdateEvent
	workerDone  chan interface{}
	numWorkers  int32
}

// Connect connects to the bus and starts listening for events exposed in the contracts.UserEvents return or an error
func (r *UMBMessageBusRepository) Connect() (evts contracts.UserEvents, err error) {
	ctx := context.Background()

	caCert, err := x509.SystemCertPool()
	if err != nil {
		return
	}

	cert, err := tls.LoadX509KeyPair(r.config.UMBClientCertFile, r.config.UMBClientCertKey)
	if err != nil {
		return
	}

	tlsConf := &tls.Config{
		RootCAs:      caCert,
		Certificates: []tls.Certificate{cert},
	}

	for {
		subCtx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(r.config.ConnectTimeoutSeconds))
		r.conn, err = amqp.Dial(subCtx, r.config.URL, &amqp.ConnOptions{
			TLSConfig: tlsConf,
		})

		cancel()

		if err == nil {
			break
		}

		glog.Errorf("Error connecting to UMB: %+v. Data may desynchronize! Retrying...", err)
		time.Sleep(time.Second * time.Duration(r.config.RetryBackoffSeconds))
	}

	// open a session
	session, err := r.conn.NewSession(ctx, nil)

	if err != nil {
		return
	}

	r.recvCtx, r.recvCancel = context.WithCancel(context.Background())
	r.changes, err = r.receiveSubjectChanges(session)
	if err != nil {
		return
	}

	return contracts.UserEvents{
		SubjectChanges: r.changes,
		Errors:         r.errs,
	}, nil
}

func (r *UMBMessageBusRepository) reconnect() {
	glog.Info("UMB attempting to reconnect.")
	r.repeatableDisconnect()
	_, err := r.Connect()
	if err != nil {
		glog.Errorf("UMB reconnect failed permanently: %s - data may desynchronize.", err)
	}
	glog.Info("UMB successfully reconnected!")
}

func (r *UMBMessageBusRepository) receiveSubjectChanges(s *amqp.Session) (chan contracts.SubjectAddOrUpdateEvent, error) {
	ctx := context.Background()
	var err error
	// create a receiver
	r.subjectRecv, err = s.NewReceiver(ctx, r.config.TopicName, nil)
	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&r.numWorkers, 1) //Atomic increment, could be modified by other goroutines in the future
	go func() {
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			err := r.subjectRecv.Close(ctx) //TODO: close correctly? Do we need another channel?
			if err != nil {
				glog.Errorf("Failed to close reciever: %v", err)
			}
			cancel()
			r.workerDone <- struct{}{}
		}()
		for {
			// receive next message
			msg, err := r.subjectRecv.Receive(r.recvCtx, nil)
			if err != nil {
				if err == context.Canceled {
					return
				}

				glog.Errorf("Reading message from AMQP: %+v", err)
				r.errs <- err

				if isConnectivityError(err) {
					go r.reconnect()
					return
				}
			}

			var evt SubjectEventMessage
			body, ok := msg.Value.(string)
			if !ok {
				glog.Errorf("Failure casting string payload to string")
			}

			err = xml.Unmarshal([]byte(body), &evt)
			if err != nil {
				r.errs <- err

				err = r.subjectRecv.RejectMessage(ctx, msg, nil)
				if err != nil {
					r.errs <- err
				}
				continue
			}

			glog.Infof("Message received. Unmarshalled Payload: %+v", evt)

			if evt.OrgID() == "" {
				r.errs <- fmt.Errorf("Unable to extract orgID from subject event. SubjectID: %s, IsUpdate: %t", evt.SubjectID(), evt.IsActive())

				err = r.subjectRecv.RejectMessage(ctx, msg, nil)
				if err != nil {
					r.errs <- err
				}
				continue
			}

			r.changes <- contracts.SubjectAddOrUpdateEvent{
				MsgRef:    msg,
				SubjectID: evt.SubjectID(),
				OrgID:     evt.OrgID(),
				Active:    evt.IsActive(),
			}
		}
	}()

	return r.changes, nil
}

// ReportSuccess sends confirmation to the broker that the message was processed successfully. This or ReportFailure MUST be called for any event received.
func (r *UMBMessageBusRepository) ReportSuccess(evt contracts.SubjectAddOrUpdateEvent) error {
	msg, ok := evt.MsgRef.(*amqp.Message)
	if ok {
		err := r.subjectRecv.AcceptMessage(context.TODO(), msg)
		return err
	}

	return fmt.Errorf("Internal error. MsgRef is not of expected type: %+v", evt.MsgRef)
}

// ReportFailure informs the broker that the message was -not- processed successfully. This or ReportSuccess MUST be called for any event received.
func (r *UMBMessageBusRepository) ReportFailure(evt contracts.SubjectAddOrUpdateEvent) error {
	msg, ok := evt.MsgRef.(*amqp.Message)
	if ok {
		err := r.subjectRecv.ReleaseMessage(context.TODO(), msg)
		return err
	}

	return fmt.Errorf("Internal error. MsgRef is not of expected type: %+v", evt.MsgRef)
}

func isConnectivityError(err error) bool {
	var connErr *amqp.ConnError
	var linkErr *amqp.LinkError
	var sessionErr *amqp.SessionError

	return errors.As(err, &connErr) ||
		errors.As(err, &linkErr) ||
		errors.As(err, &sessionErr)
}

// Disconnect disconnects from the message bus and frees any resources used for communication.
func (r *UMBMessageBusRepository) Disconnect() {
	r.repeatableDisconnect()

	close(r.errs)
	close(r.changes)
}

func (r *UMBMessageBusRepository) repeatableDisconnect() {
	r.recvCancel()
	for r.numWorkers > 0 {
		<-r.workerDone
		r.numWorkers--
	}

	err := r.conn.Close()
	if err != nil {
		glog.Errorf("Error disconnecting from AMQP broker: %s", err)
	}
}

// NewUMBMessageBusRepository constructs a new UMBMessageBusRepository with the given configuration
func NewUMBMessageBusRepository(config serviceconfig.UMBConfig) *UMBMessageBusRepository {
	return &UMBMessageBusRepository{
		config:     config,
		workerDone: make(chan interface{}),
		errs:       make(chan error),
		changes:    make(chan contracts.SubjectAddOrUpdateEvent),
	}
}
