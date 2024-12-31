package config

import (
	"os"

	"sigs.k8s.io/yaml"
)

const (
	Path    = "/etc/kubernetes/gitlab-authn.yaml"
	WebPath = "/usr/share/kubernetes-gitlab-authn/public"
)

type Config struct {
	Realms  Realms   `json:"realms"`
	Gitlab  *Gitlab  `json:"gitlab"`
	Server  *Server  `json:"server"`
	Health  *Health  `json:"health"`
	Metrics *Metrics `json:"metrics"`
	Cache   *Cache   `json:"cache"`
	Web     *Web     `json:"web"`

	file string `json:"-"`
}

func New() *Config {
	result := &Config{
		Realms:  NewRealms(),
		Gitlab:  NewGitlab(),
		Server:  NewServer(),
		Health:  NewHealth(),
		Metrics: NewMetrics(),
		Cache:   NewCache(),
		Web:     NewWeb(),
		file:    Path,
	}

	return result
}

// String returns the filesystem location
// of the file containing the stored data.
// It satisfies the [flag.Value] contract.
func (c *Config) String() string {
	return c.file
}

// Set is an alias for [Config.LoadFile].
// It satisfies the [flag.Value] contract.
func (c *Config) Set(path string) error {
	return c.LoadFile(path)
}

func (c *Config) LoadFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	c.file = path

	return yaml.Unmarshal(b, c)
}
