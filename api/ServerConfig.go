// Package api is for communication purposes
package api

const (
	// SpicedbImage is the image used for containerized spiceDB in tests
	SpicedbImage = "authzed/spicedb"
	// SpicedbVersion is the image version used for containerized spiceDB in tests
	SpicedbVersion = "v1.20.0"
)

// ServerConfig contains all server-related configuration.
type ServerConfig struct {
	GrpcPort    string
	HTTPPort    string
	HTTPSPort   string
	TLSConfig   TLSConfig
	StoreConfig StoreConfig
	AuthConfig  AuthConfig
}

// TLSConfig includes a possible TLS configuration.
type TLSConfig struct {
	CertPath string
	CertName string //default: tls.crt
	KeyPath  string
	KeyName  string //default: tls.key
}

// StoreConfig includes data used to connect to persistent storage
type StoreConfig struct {
	Store     string
	Endpoint  string
	AuthToken string
	UseTLS    bool
}

// AuthConfig holds configuration values for the oAuth client authorization middleware
type AuthConfig struct {
	DiscoveryEndpoint string
	Audience          string
	RequiredScope     string
}
