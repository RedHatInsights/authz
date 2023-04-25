// Package serviceconfig contains the typed configuration for different parts of the application.
package serviceconfig

const (
	// SpicedbImage is the image used for containerized spiceDB in tests
	SpicedbImage = "authzed/spicedb"
	// SpicedbVersion is the image version used for containerized spiceDB in tests
	SpicedbVersion = "v1.20.0"
)

// ServiceConfig contains all server-related configuration.
type ServiceConfig struct {
	GrpcPort    string
	HTTPPort    string
	HTTPSPort   string
	CorsConfig  CorsConfig
	TLSConfig   TLSConfig
	StoreConfig StoreConfig
	LogRequests bool
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
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	AllowCredentials bool
	MaxAge           int
	Debug            bool
}
