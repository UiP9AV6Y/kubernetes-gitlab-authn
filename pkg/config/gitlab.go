package config

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"
	"strings"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/access"
)

// https://docs.gitlab.com/ee/security/tokens/#token-prefixes
var GitlabTokenPrefixes = []string{
	"glpat-",
	"gloas-",
	"glsoat-",
}

type GitlabGroupFilter struct {
	OwnedOnly      bool                    `json:"owned_only"`
	TopLevelOnly   bool                    `json:"top_level_only"`
	MinAccessLevel gitlab.AccessLevelValue `json:"min_access_level"`
	Name           string                  `json:"name"`
	Limit          uint8                   `json:"limit"`
}

func (f *GitlabGroupFilter) ListOptions() *gitlab.ListGroupsOptions {
	list := gitlab.ListOptions{
		Page: 1,
	}
	result := &gitlab.ListGroupsOptions{
		ListOptions: list,
	}

	if f.Name != "" {
		result.Search = &f.Name
	}

	if f.OwnedOnly {
		result.Owned = &f.OwnedOnly
	}

	if f.TopLevelOnly {
		result.TopLevelOnly = &f.TopLevelOnly
	}

	if f.Limit > 0 {
		result.ListOptions.PerPage = int(f.Limit)
	}

	if f.MinAccessLevel > gitlab.MinimalAccessPermissions {
		result.MinAccessLevel = &f.MinAccessLevel
	}

	return result
}

type Gitlab struct {
	Server `json:",inline"`

	AttributesAsGroups bool              `json:"attributes_as_groups"`
	InactivityTimeout  Duration          `json:"inactivity_timeout"`
	GroupFilter        GitlabGroupFilter `json:"group_filter"`

	TokenPrefixes []string `json:"token_prefixes"`
}

func NewGitlab() *Gitlab {
	result := &Gitlab{
		Server:            *NewServer(),
		TokenPrefixes:     GitlabTokenPrefixes,
		InactivityTimeout: Duration{time.Hour * 24 * 30 * 6}, // ~6 months
	}
	result.Server.Address = "gitlab.com"
	result.Server.Port = 443
	result.Server.TLS = &TLS{}
	result.GroupFilter.Limit = 20                                       // Gitlab Groups API default
	result.GroupFilter.MinAccessLevel = gitlab.MinimalAccessPermissions // no filter

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

func (g *Gitlab) UserInfoOptions() *access.UserInfoOptions {
	result := &access.UserInfoOptions{
		AttributesAsGroups: g.AttributesAsGroups,
		DormantTimeout:     g.InactivityTimeout.Duration,
	}

	return result
}

func (g *Gitlab) TokenValidator() func(string) bool {
	result := func(v string) bool {
		for _, p := range g.TokenPrefixes {
			if strings.HasPrefix(v, p) {
				return true
			}
		}

		return false
	}

	return result
}
