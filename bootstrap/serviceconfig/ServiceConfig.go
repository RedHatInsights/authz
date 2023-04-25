// Package serviceconfig contains the typed configuration for different parts of the application.
package serviceconfig

// ServiceConfig contains all server-related configuration.
type ServiceConfig struct {
	GrpcPort    string
	HTTPPort    string
	HTTPSPort   string
	CorsConfig  CorsConfig
	TLSConfig   TLSConfig
	StoreConfig StoreConfig
}

// TLSConfig includes the TLS configuration.
type TLSConfig struct {
	CertFile string
	KeyFile  string
}

// StoreConfig includes connection details to use an underlying authZ store
type StoreConfig struct {
	Kind      string
	Endpoint  string
	AuthToken string
	UseTLS    bool
}

// CorsConfig includes the CORS middleware configuration
type CorsConfig struct {
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
	Debug            bool
}
