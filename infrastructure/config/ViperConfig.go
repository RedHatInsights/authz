// Package config contains the technical configuration implementations
package config

import (
	"authz/bootstrap/serviceconfig"

	"github.com/spf13/viper"
)

// ViperConfig -
type ViperConfig struct {
	v        *viper.Viper
	defaults serviceconfig.ServiceConfig
}

// Load loads the current configuration
func (c *ViperConfig) Load() (serviceconfig.ServiceConfig, error) {
	err := c.v.ReadInConfig()
	if err != nil {
		return c.defaults, err
	}

	cfg := c.defaults //Value-assignment to copy the defaults into a new object
	err = c.v.Unmarshal(&cfg)
	return cfg, err
}

// NewViperConfig constructs a new instance of ViperConfig for the given file and default values
func NewViperConfig(configFilePath string, defaults serviceconfig.ServiceConfig) *ViperConfig {
	v := viper.New()
	v.AddConfigPath(".")
	v.AddConfigPath("/")
	v.SetConfigName(configFilePath)
	v.SetConfigType("yaml")

	return &ViperConfig{
		v:        v,
		defaults: defaults,
	}
}
