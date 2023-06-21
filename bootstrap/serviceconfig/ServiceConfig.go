// Package serviceconfig contains the typed configuration for different parts of the application.
package serviceconfig

import (
	"os"
)

const (
	// SpicedbImage is the image used for containerized spiceDB in tests
	SpicedbImage = "authzed/spicedb"
	// SpicedbVersion is the image version used for containerized spiceDB in tests
	SpicedbVersion = "v1.20.0"
)

// ServiceConfig contains all server-related configuration.
type ServiceConfig struct {
	GrpcPort          int               `validate:"required,gte=0,lte=65535"`
	GrpcPortStr       string            `mapstructure:"grpcport"`
	HTTPPort          int               `validate:"required,gte=0,lte=65535"`
	HTTPPortStr       string            `mapstructure:"httpport"`
	HTTPSPort         int               `validate:"omitempty,gte=0,lte=65535"`
	HTTPSPortStr      string            `mapstructure:"httpsport"`
	CorsConfig        CorsConfig        `mapstructure:"cors"`
	TLSConfig         TLSConfig         `mapstructure:"tls"`
	StoreConfig       StoreConfig       `mapstructure:"store"`
	AuthConfigs       []AuthConfig      `mapstructure:"auth"`
	UserServiceConfig UserServiceConfig `mapstructure:"userservice"`
	LogRequests       bool
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
	TokenFile string
	UseTLS    bool
}

// ReadToken reads token from the TokenFile
func (c StoreConfig) ReadToken() (string, error) {
	bytes, err := os.ReadFile(c.TokenFile)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
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

// UserServiceConfig holds the configuration to connect to a user service API
type UserServiceConfig struct {
	URL                       string
	UserServiceClientCertFile string
	UserServiceClientKeyFile  string
	OptionalRootCA            string
	DisableCAVerification     bool
}

// UMBConfig holds the configuration to connect to the Universal Message Bus
type UMBConfig struct {
	URL               string
	UMBClientCertFile string //TODO: are these the same as the UserService files? Should they get their own credentials section or duplicate values?
	UMBClientCertKey  string
}
