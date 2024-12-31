package config

import (
	"net"
	"net/url"
	"strconv"
)

var rootPath, _ = url.Parse("/")

type TLS struct {
	SkipVerify bool `json:"skip_verify"`

	CACertFile string `json:"ca_cert_file"`
	CertFile   string `json:"cert_file"`
	KeyFile    string `json:"key_file"`
}

type Server struct {
	*TLS `json:"tls"`

	Address string `json:"address"`
	Port    uint   `json:"port"`

	Path string `json:"path"`
}

func NewServer() *Server {
	result := &Server{
	}

	return result
}

func (s *Server) HandlerPath(p string) string {
	return rootPath.JoinPath(s.Path, p).Path
}

func (s *Server) Addr() string {
	port := strconv.FormatUint(uint64(s.Port), 10)

	return net.JoinHostPort(s.Address, port)
}

func (s *Server) URL() (*url.URL, error) {
	scheme := "http"
	if s.TLS != nil {
		scheme = "https"
	}

	u, err := url.Parse(scheme + "://" + s.Addr())
	if err != nil {
		return nil, err
	}

	return u.JoinPath(s.Path), nil
}
