// Package messaging contains repository implementations for exchanging events in an enterprise environment
package messaging

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
	"context"
	"github.com/Azure/go-amqp"
	"github.com/golang/glog"
	"log"
	"strconv"
	"time"
)

var cnt = 0

// UMBMessageBusRepository can send and receive events on the Universal Message Bus
type UMBMessageBusRepository struct {
	config serviceconfig.UMBConfig
}

// Connect connects to the bus and starts listening for events exposed in the contracts.UserEvents return or an error
func (r *UMBMessageBusRepository) Connect() (contracts.UserEvents, error) {
	ctx := context.Background() // TODO: evaluate if we need a cancellable context

	conn, err := amqp.Dial(ctx, r.config.URL, &amqp.ConnOptions{
		SASLType: amqp.SASLTypePlain("reader", "password1"), //TODO: change to certs
	})

	if err != nil {
		log.Fatal("Dialing AMQP server:", err)
	}
	//defer conn.Close() // TODO: Error handling?

	// open a session
	session, err := conn.NewSession(ctx, nil)

	if err != nil {
		log.Fatal("Creating AMQP session:", err)
	}

	u, e := receiveSubjectChanges(r.config, session)
	return contracts.UserEvents{
		SubjectChanges: u,
		Errors:         e,
	}, nil
}

func receiveSubjectChanges(cfg serviceconfig.UMBConfig, s *amqp.Session) (chan contracts.SubjectAddOrUpdateEvent, chan error) {
	updates := make(chan contracts.SubjectAddOrUpdateEvent)
	errors := make(chan error)
	ctx := context.Background()
	start := time.Now()
	// create a receiver
	receiver, err := s.NewReceiver(ctx, cfg.TopicName, nil)
	if err != nil {
		glog.Errorf("Creating receiver link:", err)
		errors <- err
	}

	go func() {
		defer func() {
			ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
			receiver.Close(ctx) //TODO: close correctly? Do we need another channel?
			cancel()
		}()
		for {
			// receive next message
			msg, err := receiver.Receive(ctx, nil)
			if err != nil {
				glog.Errorf("Reading message from AMQP:", err)
				errors <- err
			}

			// accept message
			if err = receiver.AcceptMessage(context.TODO(), msg); err != nil { //TODO: switch right context
				glog.Errorf("Failure accepting message: %v", err)
				errors <- err
			}

			cnt++
			elapsed := time.Since(start).Seconds()
			glog.Errorf("Message # %s received after %f Seconds\n", strconv.Itoa(cnt), elapsed)

			updates <- contracts.SubjectAddOrUpdateEvent{
				SubjectID: "test",
				OrgID:     string(msg.GetData()), // TODO: add xml parsing. just for testing purposes now.
				Active:    false,
			}
		}
	}()

	return updates, errors
}

// Disconnect disconnects from the message bus and frees any resources used for communication.
func (r *UMBMessageBusRepository) Disconnect() {
	panic("not implemented")
}

// NewUMBMessageBusRepository constructs a new UMBMessageBusRepository with the given configuration
func NewUMBMessageBusRepository(config serviceconfig.UMBConfig) *UMBMessageBusRepository {
	return &UMBMessageBusRepository{config: config}
}
