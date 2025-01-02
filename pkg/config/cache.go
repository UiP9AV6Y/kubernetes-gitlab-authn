package config

import (
	"encoding/json"
	"fmt"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	var data interface{}
	err = json.Unmarshal(b, &data)
	if err != nil {
		return
	}

	switch value := data.(type) {
	case float64:
		d.Duration = time.Duration(value)
	case string:
		d.Duration, err = time.ParseDuration(value)
	default:
		err = fmt.Errorf("invalid duration: %#v", data)
	}

	return
}

type Cache struct {
	TTL Duration `json:"ttl"`
}

func NewCache() *Cache {
	result := &Cache{
		TTL: Duration{2 * time.Minute},
	}

	return result
}

func (c *Cache) ExpirationTime() time.Duration {
	return c.TTL.Duration
}
