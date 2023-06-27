package contracts

// SubjectAddOrUpdateEvent represents a new or updated subject in the environment
type SubjectAddOrUpdateEvent struct {
	// MsgRef represents any internal tracking information. This is meant to be used by the repository only. TODO: this wouldn't be necessary if SubjectAddOrUpdateEvent were an interface implemented by a repo-defined struct that could carry additional properties.
	MsgRef interface{}
	// SubjectID is the subject's unique id
	SubjectID string
	// OrgID is the subject's primary organization's id
	OrgID string
	// Active indicates whether or not the subject's account is active
	Active bool
}

// UserEvents represents event inputs from the environment as a set of channels
type UserEvents struct {
	// SubjectChanges events represent a new or modified subject in the environment
	SubjectChanges chan SubjectAddOrUpdateEvent
	// Errors events represent errors the repository was not able to automatically recover from after the initial connection was established
	Errors chan error
}

// MessageBusRepository represents the abstract operations for exchanging events in an enterprise environment
type MessageBusRepository interface {
	// Connect establishes a connection to the environment and, if successful, returns an UserEvents struct. If not successful, an error is returned.
	Connect() (UserEvents, error)
	// Disconnect disconnects from the environment as gracefully as possible and frees all resources allocated by Connect
	Disconnect()
	// ReportSuccess sends confirmation to the broker that the message was processed successfully. This or ReportFailure MUST be called for any event received.
	ReportSuccess(evt SubjectAddOrUpdateEvent) error
	// ReportFailure informs the broker that the message was -not- processed successfully. This or ReportSuccess MUST be called for any event received.
	ReportFailure(evt SubjectAddOrUpdateEvent) error
}
