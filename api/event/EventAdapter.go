// Package event contains adapters for dispatching events the service receives or publishes
package event

import (
	"authz/application"
	"authz/domain/contracts"

	"github.com/golang/glog"
)

// EventAdapter is used as a struct containing the services needed to process events
type EventAdapter struct {
	licenseAppService *application.LicenseAppService
	bus               contracts.MessageBusRepository
}

func NewEventAdapter(licenseAppService *application.LicenseAppService, bus contracts.MessageBusRepository) *EventAdapter {
	return &EventAdapter{
		licenseAppService: licenseAppService,
		bus:               bus,
	}
}

func (e *EventAdapter) Start() error {
	events, err := e.bus.Connect()
	if err != nil {
		return err
	}

	go e.Run(events)

	return nil
}

func (e *EventAdapter) Run(evts contracts.UserEvents) {
	ok := true
	var evt contracts.SubjectAddOrUpdateEvent
	var err error

	for ok {
		select {
		case evt, ok = <-evts.SubjectChanges:
			glog.Infof("Subject event from UMB connection: %+v", evt)
		case err, ok = <-evts.Errors:
			glog.Errorf("Error from UMB connection: %v", err)
		}
	}
}

func (e *EventAdapter) Stop() {

}
