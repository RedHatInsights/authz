// Package messaging contains repository implementations for exchanging events in an enterprise environment
package messaging

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/xml"
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
	config     serviceconfig.UMBConfig
	conn       *amqp.Conn
	recvCtx    context.Context
	recvCancel context.CancelFunc
	errs       chan error
	workerDone chan interface{}
	numWorkers int32
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

	r.conn, err = amqp.Dial(ctx, r.config.URL, &amqp.ConnOptions{
		TLSConfig: tlsConf,
	})

	if err != nil {
		return
	}

	// open a session
	session, err := r.conn.NewSession(ctx, nil)

	if err != nil {
		return
	}

	r.recvCtx, r.recvCancel = context.WithCancel(context.Background())
	r.errs = make(chan error)
	r.workerDone = make(chan interface{})
	u, err := r.receiveSubjectChanges(session)
	if err != nil {
		return
	}

	return contracts.UserEvents{
		SubjectChanges: u,
		Errors:         r.errs,
	}, nil
}

func (r *UMBMessageBusRepository) receiveSubjectChanges(s *amqp.Session) (chan contracts.SubjectAddOrUpdateEvent, error) {
	updates := make(chan contracts.SubjectAddOrUpdateEvent)
	ctx := context.Background()
	// create a receiver
	receiver, err := s.NewReceiver(ctx, r.config.TopicName, nil)
	if err != nil {
		return nil, err
	}

	atomic.AddInt32(&r.numWorkers, 1) //Atomic increment, could be modified by other goroutines in the future
	go func() {
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			err := receiver.Close(ctx) //TODO: close correctly? Do we need another channel?
			if err != nil {
				glog.Errorf("Failed to close reciever: %v", err)
			}
			cancel()
			close(updates)
			r.workerDone <- struct{}{}
		}()
		for {
			// receive next message
			msg, err := receiver.Receive(r.recvCtx, nil)
			if err != nil {
				if err == context.Canceled {
					return
				}
				glog.Errorf("Reading message from AMQP:", err)
				r.errs <- err
			}

			var evt SubjectEventMessage
			body, ok := msg.Value.(string)
			if !ok {
				glog.Errorf("Failure casting string payload to string")
			}

			err = xml.Unmarshal([]byte(body), &evt)
			if err != nil {
				r.errs <- err
				//Reject message- unparseable
				continue
			}

			glog.Infof("Message received. Unmarshalled Payload: %v", evt)

			if evt.OrgID() == "" {
				r.errs <- fmt.Errorf("Unable to extract orgID from subject event. SubjectID: %s, IsUpdate: %t", evt.SubjectID(), evt.IsActive())
				//Reject message- no orgid
				continue
			}

			// accept message - should happen after successful processing, otherwise release message
			if err = receiver.AcceptMessage(context.TODO(), msg); err != nil { //TODO: switch right context
				glog.Errorf("Failure accepting message: %v", err)
				r.errs <- err
			}

			updates <- contracts.SubjectAddOrUpdateEvent{
				SubjectID: evt.SubjectID(),
				OrgID:     evt.OrgID(),
				Active:    evt.IsActive(),
			}
		}
	}()

	return updates, nil
}

// Disconnect disconnects from the message bus and frees any resources used for communication.
func (r *UMBMessageBusRepository) Disconnect() {
	r.recvCancel()
	for r.numWorkers > 0 {
		<-r.workerDone
		r.numWorkers--
	}

	err := r.conn.Close()
	if err != nil {
		glog.Errorf("Error disconnecting from AMQP broker: %s", err)
	}

	close(r.errs)
}

// NewUMBMessageBusRepository constructs a new UMBMessageBusRepository with the given configuration
func NewUMBMessageBusRepository(config serviceconfig.UMBConfig) *UMBMessageBusRepository {
	return &UMBMessageBusRepository{config: config}
}
