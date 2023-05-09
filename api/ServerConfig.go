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
	RequiredScope     string
	// 1) config: includes DiscoveryEndpoint
	// 2) Struct that holds issuer and JWKS (+ cache handling and retry mechanism) at bootstrap of the interceptor
	// 3) middleware: use public key and issuer from discovery endpoint to validate token contents against it
	// 4) check for needed scope to access api (injected also via config) - scope is "api.iam.access"
}
