// Package api is for communication purposes
package api

// ServerConfig contains all server-related configuration.
type ServerConfig struct {
	GrpcPort  string
	HTTPPort  string
	HTTPSPort string
	TLSConfig TLSConfig
}

// TLSConfig includes a possible TLS configuration.
type TLSConfig struct {
	CertPath string
	CertName string //default: tls.crt
	KeyPath  string
	KeyName  string //default: tls.key
}
