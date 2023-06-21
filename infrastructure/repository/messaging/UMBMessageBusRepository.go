// Package messaging contains repository implementations for exchanging events in an enterprise environment
package messaging

import (
	"authz/bootstrap/serviceconfig"
	"authz/domain/contracts"
)

// UMBMessageBusRepository can send and receive events on the Universal Message Bus
type UMBMessageBusRepository struct {
	config serviceconfig.UMBConfig
}

// Connect connects to the bus and starts listening for events exposed in the contracts.Events return or an error
func (r *UMBMessageBusRepository) Connect() (contracts.Events, error) {
	panic("not implemented")
}

// Disconnect disconnects from the message bus and frees any resources used for communication.
func (r *UMBMessageBusRepository) Disconnect() {
	panic("not implemented")
}

// NewUMBMessageBusRepository constructs a new UMBMessageBusRepository with the given configuration
func NewUMBMessageBusRepository(config serviceconfig.UMBConfig) *UMBMessageBusRepository {
	return &UMBMessageBusRepository{config: config}
}
