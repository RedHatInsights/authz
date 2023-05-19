package serviceconfig

// Config defines an interface that's capable to provide config for interaction in the domain layer without
// tightly coupling to a config implementation.
type Config interface {
	Load() (ServiceConfig, error)
}
