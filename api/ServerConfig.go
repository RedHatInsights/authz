// Package api is for communication purposes
package api

// ServerConfig contains all server-related configuration.
type ServerConfig struct {
	GrpcPort    string
	HTTPPort    string
	HTTPSPort   string
	TLSConfig   TLSConfig
	StoreConfig StoreConfig
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
