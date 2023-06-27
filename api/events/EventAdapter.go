// Package events contains adapters for dispatching events the service receives or publishes
package events

import (
	"authz/application"
	"authz/domain/contracts"

	"github.com/golang/glog"
)

// EventAdapter is used as a struct containing the services needed to process events
type EventAdapter struct {
	licenseAppService *application.LicenseAppService
	bus               contracts.MessageBusRepository
	done              chan interface{}
}

// NewEventAdapter constructs a new event adapter object from the given dependencies
func NewEventAdapter(licenseAppService *application.LicenseAppService, bus contracts.MessageBusRepository) *EventAdapter {
	return &EventAdapter{
		licenseAppService: licenseAppService,
		bus:               bus,
		done:              make(chan interface{}),
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
			glog.Infof("Subject event from UMB connection: %+v", evt)
			err = e.licenseAppService.HandleSubjectAddOrUpdateEvent(evt)
			if err == nil {
				//Accept message
			} else {
				glog.Errorf("Error processing message %+v: %v", evt, err)
				//Release message
			}
		case err, ok = <-evts.Errors:
			glog.Errorf("Error from UMB connection: %v", err)
		}
	}
	e.done <- struct{}{}
}

// Stop disconnects from the message bus, completes any message processing in progress, and then returns
func (e *EventAdapter) Stop() {
	e.bus.Disconnect()
	<-e.done
}
