// Package serviceconfig contains the typed configuration for different parts of the application.
package serviceconfig

// ServiceConfig contains all server-related configuration.
type ServiceConfig struct {
	GrpcPort    string
	HttpPort    string
	HttpsPort   string
	CorsConfig  CorsConfig
	TLSConfig   TLSConfig
	StoreConfig StoreConfig
}

// TLSConfig includes the TLS configuration.
type TLSConfig struct {
	CertPath string
	CertName string //default: tls.crt
	KeyPath  string
	KeyName  string //default: tls.key
}

// StoreConfig includes connection details to use an underlying authZ store
type StoreConfig struct {
	Store     string
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
