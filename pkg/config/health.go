package config

type Health struct {
	Server `json:",inline"`
}

func NewHealth() *Health {
	result := &Health{
		Server: *NewServer(),
	}
	result.Server.Path = "-"

	return result
}
