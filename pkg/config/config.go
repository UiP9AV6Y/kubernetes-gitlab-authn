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
	Gitlab  *Gitlab  `json:"gitlab"`
	Server  *Server  `json:"server"`
	Health  *Health  `json:"health"`
	Metrics *Metrics `json:"metrics"`
	Web     *Web     `json:"web"`
}

func New() *Config {
	result := &Config{
		Gitlab:  NewGitlab(),
		Server:  NewServer(),
		Health:  NewHealth(),
		Metrics: NewMetrics(),
		Web:     NewWeb(),
	}

	return result
}

func (c *Config) LoadFile(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(b, c)
}
