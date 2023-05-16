package bootstrap

import (
	"authz/bootstrap/serviceconfig"
	"authz/infrastructure/config"

	"github.com/spf13/viper"
)

// NewBuilderT -
type NewBuilderT struct {
	cfg *config.ViperConfig
}

// CfgNameT -
type CfgNameT struct {
	cfg *config.ViperConfig
}

// CfgTypeT -
type CfgTypeT struct {
	cfg *config.ViperConfig
}

// DefaultsT -
type DefaultsT struct {
	cfg *config.ViperConfig
}

// CfgPathsT -
type CfgPathsT struct {
	cfg *config.ViperConfig
}

// OptionsT -
type OptionsT struct {
	cfg *config.ViperConfig
}

const (
	// NoErrorOnMissingCfgFileOption defines that there should be no error upon creation of Cfg if none of config files could be found.
	NoErrorOnMissingCfgFileOption serviceconfig.CfgOption = "NoErrorOnMissingCfgFileOption"
)

// NewConfigurationBuilder creates a new config builder
func NewConfigurationBuilder() *NewBuilderT {
	return &NewBuilderT{
		cfg: &config.ViperConfig{
			V:       viper.New(),
			Options: make(map[serviceconfig.CfgOption]bool),
		},
	}
}

// ConfigName - name
func (t *NewBuilderT) ConfigName(name string) *CfgNameT {
	t.cfg.V.SetConfigName(name)
	return &CfgNameT{
		cfg: t.cfg,
	}
}

// ConfigType - type (yaml, etc)
func (t *CfgNameT) ConfigType(typ string) *CfgTypeT {
	t.cfg.V.SetConfigType(typ)
	return &CfgTypeT{
		cfg: t.cfg,
	}
}

func configPaths(cfg *config.ViperConfig, paths ...string) *CfgPathsT {
	for _, p := range paths {
		cfg.V.AddConfigPath(p)
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

func defaults(c *config.ViperConfig, defs map[string]interface{}) *DefaultsT {
	for k, v := range defs {
		c.V.SetDefault(k, v)
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
func (t *DefaultsT) Options(options ...serviceconfig.CfgOption) *OptionsT {
	for _, i := range options {
		t.cfg.Options[i] = true
	}
	return &OptionsT{
		cfg: t.cfg,
	}
}

// Build - builds the config using the domain.contracts.config contract
func (t *OptionsT) Build() (serviceconfig.Config, error) {
	err := t.cfg.V.ReadInConfig()

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
