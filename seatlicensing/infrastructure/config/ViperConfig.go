package config

import (
	"authz/seatlicensing/domain/contracts"
	"time"

	"github.com/spf13/viper"
)

type ViperConfig struct {
	v       *viper.Viper
	options map[CfgOption]bool
}

type NewBuilderT struct {
	cfg *ViperConfig
}

type configNameT struct {
	cfg *ViperConfig
}

type configTypeT struct {
	cfg *ViperConfig
}

type defaultsT struct {
	cfg *ViperConfig
}

type configPathsT struct {
	cfg *ViperConfig
}

type optionsT struct {
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

func (t *NewBuilderT) ConfigName(name string) *configNameT {
	t.cfg.v.SetConfigName(name)
	return &configNameT{
		cfg: t.cfg,
	}
}

func (t *configNameT) ConfigType(typ string) *configTypeT {
	t.cfg.v.SetConfigType(typ)
	return &configTypeT{
		cfg: t.cfg,
	}
}

func configPaths(cfg *ViperConfig, paths ...string) *configPathsT {
	for _, p := range paths {
		cfg.v.AddConfigPath(p)
	}
	return &configPathsT{
		cfg: cfg,
	}
}

func (t *configNameT) ConfigPaths(path ...string) *configPathsT {
	return configPaths(t.cfg, path...)
}

func (t *configTypeT) ConfigPaths(path ...string) *configPathsT {
	return configPaths(t.cfg, path...)
}

func defaults(c *ViperConfig, defs map[string]interface{}) *defaultsT {
	for k, v := range defs {
		c.v.SetDefault(k, v)
	}
	return &defaultsT{
		cfg: c,
	}
}

func (t *configPathsT) Defaults(defs map[string]interface{}) *defaultsT {
	return defaults(t.cfg, defs)
}

func (t *configPathsT) NoDefaults() *defaultsT {
	return &defaultsT{
		cfg: t.cfg,
	}
}

func (t *defaultsT) Options(options ...CfgOption) *optionsT {
	for _, i := range options {
		t.cfg.options[i] = true
	}
	return &optionsT{
		cfg: t.cfg,
	}
}

func (c *ViperConfig) HasOption(option CfgOption) bool {
	if _, ok := c.options[option]; ok {
		return true
	}
	return false
}

func (t *optionsT) Build() (contracts.Config, error) {
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

func (c *ViperConfig) GetInt(key string) int {
	return c.v.GetInt(key)
}

func (c *ViperConfig) GetInt32(key string) int32 {
	return c.v.GetInt32(key)
}

func (c *ViperConfig) GetInt64(key string) int64 {
	return c.v.GetInt64(key)
}

func (c *ViperConfig) GetUint(key string) uint {
	return c.v.GetUint(key)
}

func (c *ViperConfig) GetUint32(key string) uint32 {
	return c.v.GetUint32(key)
}

func (c *ViperConfig) GetUint64(key string) uint64 {
	return c.v.GetUint64(key)
}

func (c *ViperConfig) GetString(key string) string {
	return c.v.GetString(key)
}

func (c *ViperConfig) GetBool(key string) bool {
	return c.v.GetBool(key)
}

func (c *ViperConfig) Get(key string) interface{} {
	return c.v.Get(key)
}

func (c *ViperConfig) GetFloat64(key string) float64 {
	return c.v.GetFloat64(key)
}

func (c *ViperConfig) GetAll() map[string]interface{} {
	return c.v.AllSettings()
}

func (c *ViperConfig) GetDuration(key string) time.Duration {
	return c.v.GetDuration(key)
}

func (c *ViperConfig) GetIntSlice(key string) []int {
	return c.v.GetIntSlice(key)
}

func (c *ViperConfig) GetSizeInBytes(key string) uint {
	return c.v.GetSizeInBytes(key)
}

func (c *ViperConfig) GetTime(key string) time.Time {
	return c.v.GetTime(key)
}

func (c *ViperConfig) GetStringSlice(key string) []string {
	return c.v.GetStringSlice(key)
}

func (c *ViperConfig) Set(key string, val interface{}) {
	c.v.Set(key, val)
}

func (c *ViperConfig) Write() error {
	return c.v.WriteConfig()
}
