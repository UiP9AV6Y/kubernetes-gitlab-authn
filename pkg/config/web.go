package config

type Web struct {
	Path        string `json:"path"`
	Description string `json:"description"`
}

func NewWeb() *Web {
	result := &Web{
		Path: WebPath,
	}

	return result
}
