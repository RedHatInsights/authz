// Package config contains the typed configuration for different parts of the application.
package config

// ServerConfig contains all server-related configuration.
type ServerConfig struct {
	Kind             string
	MainPort         string
	GrpcWebHttpPort  string
	GrpcWebHttpsPort string
	TlsConfig        TlsConfig
}

type TlsConfig struct {
	CertPath string
	CertName string //default: tls.crt
	KeyPath  string
	KeyName  string //default: tls.key
}
