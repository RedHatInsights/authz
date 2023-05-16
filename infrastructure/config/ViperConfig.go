// Package config contains the technical configuration implementations
package config

import (
	"authz/bootstrap/serviceconfig"
	"time"

	"github.com/spf13/viper"
)

// ViperConfig -
type ViperConfig struct {
	V       *viper.Viper
	Options map[serviceconfig.CfgOption]bool
}

// HasOption -
func (c *ViperConfig) HasOption(option serviceconfig.CfgOption) bool {
	if _, ok := c.Options[option]; ok {
		return true
	}
	return false
}

// GetInt -
func (c *ViperConfig) GetInt(key string) int {
	return c.V.GetInt(key)
}

// GetInt32 -
func (c *ViperConfig) GetInt32(key string) int32 {
	return c.V.GetInt32(key)
}

// GetInt64 -
func (c *ViperConfig) GetInt64(key string) int64 {
	return c.V.GetInt64(key)
}

// GetUint -
func (c *ViperConfig) GetUint(key string) uint {
	return c.V.GetUint(key)
}

// GetUint32 -
func (c *ViperConfig) GetUint32(key string) uint32 {
	return c.V.GetUint32(key)
}

// GetUint64 -
func (c *ViperConfig) GetUint64(key string) uint64 {
	return c.V.GetUint64(key)
}

// GetString -
func (c *ViperConfig) GetString(key string) string {
	return c.V.GetString(key)
}

// GetBool -
func (c *ViperConfig) GetBool(key string) bool {
	return c.V.GetBool(key)
}

// Get -
func (c *ViperConfig) Get(key string) interface{} {
	return c.V.Get(key)
}

// GetFloat64 -
func (c *ViperConfig) GetFloat64(key string) float64 {
	return c.V.GetFloat64(key)
}

// GetAll get all config
func (c *ViperConfig) GetAll() map[string]interface{} {
	return c.V.AllSettings()
}

// GetDuration -
func (c *ViperConfig) GetDuration(key string) time.Duration {
	return c.V.GetDuration(key)
}

// GetIntSlice -
func (c *ViperConfig) GetIntSlice(key string) []int {
	return c.V.GetIntSlice(key)
}

// GetSizeInBytes -
func (c *ViperConfig) GetSizeInBytes(key string) uint {
	return c.V.GetSizeInBytes(key)
}

// GetTime -
func (c *ViperConfig) GetTime(key string) time.Time {
	return c.V.GetTime(key)
}

// GetStringSlice -
func (c *ViperConfig) GetStringSlice(key string) []string {
	return c.V.GetStringSlice(key)
}

// Set -
func (c *ViperConfig) Set(key string, val interface{}) {
	c.V.Set(key, val)
}

func (c *ViperConfig) Write() error {
	return c.V.WriteConfig()
}
