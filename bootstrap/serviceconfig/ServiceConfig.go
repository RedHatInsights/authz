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
	GrpcPort     int `validate:"required,gte=0,lte=65535"`
	GrpcPortStr  string
	HTTPPort     int `validate:"required,gte=0,lte=65535"`
	HTTPPortStr  string
	HTTPSPort    int `validate:"omitempty,gte=0,lte=65535"`
	HTTPSPortStr string
	CorsConfig   CorsConfig
	TLSConfig    TLSConfig
	StoreConfig  StoreConfig
	AuthConfig   AuthConfig
	LogRequests  bool
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

// AuthConfig holds the configuration for the client authz middleware
type AuthConfig struct {
	DiscoveryEndpoint string
	Audience          string
	RequiredScope     string
	Enabled           bool
}
