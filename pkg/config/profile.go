package config

type Profile struct {
	Server `json:",inline"`
}

func NewProfile() *Profile {
	result := &Profile{
		Server: *NewServer(),
	}
	result.Server.Path = "debug"

	return result
}
