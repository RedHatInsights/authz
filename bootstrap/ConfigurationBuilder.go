package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/infrastructure/config"
)

// ConfigurationBuilder -
type ConfigurationBuilder struct {
	filePath string
	defaults serviceconfig.ServiceConfig
}

// NewConfigurationBuilder creates a new config builder
func NewConfigurationBuilder() *ConfigurationBuilder {
	return &ConfigurationBuilder{}
}

// ConfigFilePath sets the relative or absolute path to the configuration file.
func (t *ConfigurationBuilder) ConfigFilePath(configFilePath string) *ConfigurationBuilder {
	t.filePath = configFilePath
	return t
}

// Defaults add static defaults
func (t *ConfigurationBuilder) Defaults(defaults serviceconfig.ServiceConfig) *ConfigurationBuilder {
	t.defaults = defaults
	return t
}

// NoDefaults -
func (t *ConfigurationBuilder) NoDefaults() *ConfigurationBuilder {
	t.defaults = serviceconfig.ServiceConfig{}
	return t
}

// Build - builds the config using the domain.contracts.config contract
func (t *ConfigurationBuilder) Build() (serviceconfig.Config, error) {
	return config.NewViperConfig(t.filePath, t.defaults), nil
}
