package config

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"
)

type GitlabGroupFilter struct {
	OwnedOnly      bool                    `json:"owned_only"`
	TopLevelOnly   bool                    `json:"top_level_only"`
	MinAccessLevel gitlab.AccessLevelValue `json:"min_access_level"`
	Name           string                  `json:"name"`
}

type Gitlab struct {
	Server `json:",inline"`

	AttributesAsGroups bool              `json:"attributes_as_groups"`
	InactivityTimeout  time.Duration     `json:"inactivity_timeout"`
	GroupFilter        GitlabGroupFilter `json:"group_filter"`
}

func NewGitlab() *Gitlab {
	result := &Gitlab{
		Server:            *NewServer(),
		InactivityTimeout: time.Hour * 24 * 30 * 6, // ~6 months
	}
	result.Server.Address = "gitlab.com"
	result.Server.Port = 443
	result.Server.TLS = &TLS{}
	result.GroupFilter.MinAccessLevel = gitlab.MinimalAccessPermissions

	return result
}

func (g *Gitlab) HTTPClient() (client *http.Client, err error) {
	client = http.DefaultClient
	transport, err := g.HTTPTransport()
	if err != nil {
		return
	}

	if transport == http.DefaultTransport {
		return
	}

	client = &http.Client{
		Transport: transport,
	}
	return

}

func (g *Gitlab) HTTPTransport() (transport http.RoundTripper, err error) {
	transport = http.DefaultTransport
	mtls, err := g.MTLS()
	if err != nil {
		return
	}

	if mtls == nil {
		return
	}

	transport = &http.Transport{
		TLSClientConfig: mtls,
	}
	return
}

func (g *Gitlab) MTLS() (cfg *tls.Config, err error) {
	var pool *x509.CertPool
	var certs []tls.Certificate
	var skipVerify bool

	pool, err = g.CertPool()
	if err != nil {
		return
	}

	certs, err = g.Certificates()
	if err != nil {
		return
	}

	if g.Server.TLS != nil {
		skipVerify = g.Server.TLS.SkipVerify
	}

	if pool == nil && certs == nil && !skipVerify {
		return nil, nil
	}

	cfg = &tls.Config{
		RootCAs:            pool,
		Certificates:       certs,
		InsecureSkipVerify: skipVerify,
	}
	return
}

func (g *Gitlab) CertPool() (pool *x509.CertPool, err error) {
	var cert []byte
	if g.Server.TLS == nil || g.Server.TLS.CACertFile == "" {
		return nil, nil
	}

	cert, err = os.ReadFile(g.CACertFile)
	if err != nil {
		return nil, err
	}

	pool = x509.NewCertPool()
	pool.AppendCertsFromPEM(cert)
	return
}

func (g *Gitlab) Certificates() (certs []tls.Certificate, err error) {
	var certificate tls.Certificate
	if g.Server.TLS == nil || g.Server.CertFile == "" || g.Server.KeyFile == "" {
		return nil, nil
	}

	certificate, err = tls.LoadX509KeyPair(g.Server.TLS.CertFile, g.Server.TLS.KeyFile)
	if err != nil {
		return nil, err
	}

	certs = []tls.Certificate{certificate}
	return
}
