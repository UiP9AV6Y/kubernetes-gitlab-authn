package stdflags

import (
	"flag"
	"os"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
)

const (
  // Usage text for -config
	FlagUsageConfig  = "Configuration file location"
)

const (
  // Flag name for the config filepath
	FlagConfig  = "config"
)

const (
  // Environment variable for -config
	EnvConfig  = "CONFIG"
)

// ConfigFlags is a data storage for [flag.FlagSet] parsing results.
type ConfigFlags struct {
	cfgValue   *config.Config
	cfgEnv     string
	fs         *flag.FlagSet
}

func newConfigFlags(fs *flag.FlagSet) *ConfigFlags {
	result := &ConfigFlags{
		cfgValue:   config.New(),
		cfgEnv:     EnvConfig,
		fs:         fs,
	}

	return result
}

// NewConfigFlags returns a [ConfigFlags] instance with the given
// flagset primed for populating its internal state.
func NewConfigFlags(fs *flag.FlagSet) *ConfigFlags {
	result := newConfigFlags(fs)

	fs.Var(result.cfgValue, FlagConfig, FlagUsageConfig)

	return result
}

// NewEnvConfigFlags returns a [ConfigFlags] instance with the given
// flagset primed for populating its internal state. The flags usage
// description will include mentions of environment variables.
// If this is not desired, use [NewConfigFlags].
//
// The optional prefix is used as-is to create the lookup key for
// [ConfigFlags.ParseFunc].
func NewEnvConfigFlags(fs *flag.FlagSet, prefix string) *ConfigFlags {
	result := newConfigFlags(fs)
	result.cfgEnv = prefix + result.cfgEnv

	fs.Var(result.cfgValue, FlagConfig, FlagUsageConfig+" [$"+result.cfgEnv+"]")

	return result
}

// ParseFunc uses the provided function to retrieve values
// for the previously provisioned flagset. Values are only
// forwarded if they are not empty. Returned errors originate
// from the flag parsing logic.
func (f *ConfigFlags) ParseFunc(get func(string) string) error {
	if cfg := get(f.cfgEnv); cfg != "" {
		if err := f.fs.Set(FlagConfig, cfg); err != nil {
			return err
		}
	}

	return nil
}

// ParseEnv calls [ConfigFlags.ParseFunc] with [os.Getenv].
func (f *ConfigFlags) ParseEnv() error {
	return f.ParseFunc(os.Getenv)
}

// Config returns the parsed config.
func (f *ConfigFlags) Config() *config.Config {
	return f.cfgValue
}
