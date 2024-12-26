package config

import (
	"time"
)

type Metrics struct {
	Server `json:",inline"`

	RequestLimit   int           `json:"request_limit"`
	RequestTimeout time.Duration `json:"request_timeout"`
}

func NewMetrics() *Metrics {
	result := &Metrics{
		Server: *NewServer(),
	}
	result.Server.Path = "metrics"

	return result
}
