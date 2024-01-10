// Package events contains adapters for dispatching events the service receives or publishes
package events

import (
	"authz/domain/contracts"

	"github.com/golang/glog"
)

// EventAdapter is used as a struct containing the services needed to process events
type EventAdapter struct {
	bus  contracts.MessageBusRepository
	done chan interface{}
}

// NewEventAdapter constructs a new event adapter object from the given dependencies
func NewEventAdapter(bus contracts.MessageBusRepository) *EventAdapter {
	return &EventAdapter{
		bus:  bus,
		done: make(chan interface{}),
	}
}

// Start connects to the message bus and begins event processing
func (e *EventAdapter) Start() error {
	events, err := e.bus.Connect()
	if err != nil {
		return err
	}

	go e.run(events)

	return nil
}

func (e *EventAdapter) run(evts contracts.UserEvents) {
	ok := true
	var evt contracts.SubjectAddOrUpdateEvent
	var err error

	for ok {
		select {
		case evt, ok = <-evts.SubjectChanges:
			if ok {
				glog.Infof("Subject event from UMB connection: %+v", evt)
				e.sendResult(evt, err)
			}
		case err, ok = <-evts.Errors:
			if ok {
				glog.Errorf("Error from UMB connection: %v", err)
			}
		}
	}
	e.done <- struct{}{}
}

func (e *EventAdapter) sendResult(evt contracts.SubjectAddOrUpdateEvent, err error) {
	if err == nil {
		err = e.bus.ReportSuccess(evt)

		if err != nil {
			glog.Errorf("Error reporting success: %v", err)
		}
	} else {
		glog.Errorf("Error processing message %+v: %v", evt, err)
		err = e.bus.ReportFailure(evt)

		if err != nil {
			glog.Errorf("Error reporting failure: %v", err)
		}
	}
}

// Stop disconnects from the message bus, completes any message processing in progress, and then returns
func (e *EventAdapter) Stop() {
	e.bus.Disconnect()
	<-e.done
}
