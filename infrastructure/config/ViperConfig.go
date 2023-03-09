package config

import (
	"authz/domain/contracts"
	"time"

	"github.com/spf13/viper"
)

// ViperConfig -
type ViperConfig struct {
	v       *viper.Viper
	options map[CfgOption]bool
}

// NewBuilderT -
type NewBuilderT struct {
	cfg *ViperConfig
}

// CfgNameT -
type CfgNameT struct {
	cfg *ViperConfig
}

// CfgTypeT -
type CfgTypeT struct {
	cfg *ViperConfig
}

// DefaultsT -
type DefaultsT struct {
	cfg *ViperConfig
}

// CfgPathsT -
type CfgPathsT struct {
	cfg *ViperConfig
}

// OptionsT -
type OptionsT struct {
	cfg *ViperConfig
}

const (
	// NoErrorOnMissingCfgFileOption defines that there should be no error upon creation of Cfg if none of config files could be found.
	NoErrorOnMissingCfgFileOption CfgOption = "NoErrorOnMissingCfgFileOption"
)

// NewBuilder creates a new config builder
func NewBuilder() *NewBuilderT {
	return &NewBuilderT{
		cfg: &ViperConfig{
			v:       viper.New(),
			options: make(map[CfgOption]bool),
		},
	}
}

// ConfigName - name
func (t *NewBuilderT) ConfigName(name string) *CfgNameT {
	t.cfg.v.SetConfigName(name)
	return &CfgNameT{
		cfg: t.cfg,
	}
}

// ConfigType - type (yaml, etc)
func (t *CfgNameT) ConfigType(typ string) *CfgTypeT {
	t.cfg.v.SetConfigType(typ)
	return &CfgTypeT{
		cfg: t.cfg,
	}
}

func configPaths(cfg *ViperConfig, paths ...string) *CfgPathsT {
	for _, p := range paths {
		cfg.v.AddConfigPath(p)
	}
	return &CfgPathsT{
		cfg: cfg,
	}
}

// ConfigPaths -
func (t *CfgNameT) ConfigPaths(path ...string) *CfgPathsT {
	return configPaths(t.cfg, path...)
}

// ConfigPaths -
func (t *CfgTypeT) ConfigPaths(path ...string) *CfgPathsT {
	return configPaths(t.cfg, path...)
}

func defaults(c *ViperConfig, defs map[string]interface{}) *DefaultsT {
	for k, v := range defs {
		c.v.SetDefault(k, v)
	}
	return &DefaultsT{
		cfg: c,
	}
}

// Defaults add static defaults
func (t *CfgPathsT) Defaults(defs map[string]interface{}) *DefaultsT {
	return defaults(t.cfg, defs)
}

// NoDefaults -
func (t *CfgPathsT) NoDefaults() *DefaultsT {
	return &DefaultsT{
		cfg: t.cfg,
	}
}

// Options -
func (t *DefaultsT) Options(options ...CfgOption) *OptionsT {
	for _, i := range options {
		t.cfg.options[i] = true
	}
	return &OptionsT{
		cfg: t.cfg,
	}
}

// HasOption -
func (c *ViperConfig) HasOption(option CfgOption) bool {
	if _, ok := c.options[option]; ok {
		return true
	}
	return false
}

// Build - builds the config using the domain.contracts.config contract
func (t *OptionsT) Build() (contracts.Config, error) {
	err := t.cfg.v.ReadInConfig()

	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			if !t.cfg.HasOption(NoErrorOnMissingCfgFileOption) {
				return nil, err
			}
		} else {
			return nil, err
		}
	}
	return t.cfg, nil
}

// GetInt -
func (c *ViperConfig) GetInt(key string) int {
	return c.v.GetInt(key)
}

// GetInt32 -
func (c *ViperConfig) GetInt32(key string) int32 {
	return c.v.GetInt32(key)
}

// GetInt64 -
func (c *ViperConfig) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

// GetUint -
func (c *ViperConfig) GetUint(key string) uint {
	return c.v.GetUint(key)
}

// GetUint32 -
func (c *ViperConfig) GetUint32(key string) uint32 {
	return c.v.GetUint32(key)
}

// GetUint64 -
func (c *ViperConfig) GetUint64(key string) uint64 {
	return c.v.GetUint64(key)
}

// GetString -
func (c *ViperConfig) GetString(key string) string {
	return c.v.GetString(key)
}

// GetBool -
func (c *ViperConfig) GetBool(key string) bool {
	return c.v.GetBool(key)
}

// Get -
func (c *ViperConfig) Get(key string) interface{} {
	return c.v.Get(key)
}

// GetFloat64 -
func (c *ViperConfig) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

// GetAll get all config
func (c *ViperConfig) GetAll() map[string]interface{} {
	return c.v.AllSettings()
}

// GetDuration -
func (c *ViperConfig) GetDuration(key string) time.Duration {
	return c.v.GetDuration(key)
}

// GetIntSlice -
func (c *ViperConfig) GetIntSlice(key string) []int {
	return c.v.GetIntSlice(key)
}

// GetSizeInBytes -
func (c *ViperConfig) GetSizeInBytes(key string) uint {
	return c.v.GetSizeInBytes(key)
}

// GetTime -
func (c *ViperConfig) GetTime(key string) time.Time {
	return c.v.GetTime(key)
}

// GetStringSlice -
func (c *ViperConfig) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

// Set -
func (c *ViperConfig) Set(key string, val interface{}) {
	c.v.Set(key, val)
}

func (c *ViperConfig) Write() error {
	return c.v.WriteConfig()
}
