package config

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"os"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/access"
)

type GitlabAccessRules struct {
	// Reject users without 2FA set up
	Require2FA bool `json:"require_2fa"`
	// Reject users marked as robots
	RejectBots bool `json:"reject_bots"`
	// Reject users in locked state
	RejectLocked bool `json:"reject_locked"`
	// Reject users which have not confirmed their account yet
	RejectPristine bool `json:"reject_pristine"`
	// Only allow users with the given usernames
	RequireUsers []string `json:"require_users"`
	// Reject users based on their username
	RejectUsers []string `json:"reject_users"`
	// Require membership of all of these groups
	RequireGroups []string `json:"require_groups"`
	// Reject members of any of the given groups
	RejectGroups []string `json:"reject_groups"`
}

func (r *GitlabAccessRules) UserRules() access.AccessRuler {
	result := []access.AccessRuler{}
	if r == nil {
		return access.AllAccessRulers(result)
	}

	if r.Require2FA {
		result = append(result, access.User2FARequirement)
	}

	if r.RejectBots {
		result = append(result, access.UserBotRejection)
	}

	if r.RejectLocked {
		result = append(result, access.UserLockedRejection)
	}

	if r.RejectPristine {
		result = append(result, access.UserPristineRejection)
	}

	if r.RequireUsers != nil {
		result = append(result, access.UserNameRequirement(r.RequireUsers))
	}

	if r.RejectUsers != nil {
		result = append(result, access.UserNameRejection(r.RejectUsers))
	}

	if r.RequireGroups != nil {
		result = append(result, access.AllGroupNameRequirement(r.RejectUsers))
	}

	if r.RejectGroups != nil {
		result = append(result, access.AnyGroupNameRejection(r.RejectUsers))
	}

	return access.AllAccessRulers(result)
}

type GitlabAccessMultiRules []*GitlabAccessRules

func (r GitlabAccessMultiRules) UserRules() access.AccessRuler {
	result := make([]access.AccessRuler, len(r))
	for i, u := range r {
		result[i] = u.UserRules()
	}

	return access.AnyAccessRulers(result)
}

type GitlabGroupFilter struct {
	OwnedOnly      bool                    `json:"owned_only"`
	TopLevelOnly   bool                    `json:"top_level_only"`
	MinAccessLevel gitlab.AccessLevelValue `json:"min_access_level"`
	Name           string                  `json:"name"`
}

type Gitlab struct {
	Server `json:",inline"`

	AttributesAsGroups bool                              `json:"attributes_as_groups"`
	GroupFilter        GitlabGroupFilter                 `json:"group_filter"`
	RealmACLs          map[string]GitlabAccessMultiRules `json:"realm_acls"`
}

func NewGitlab() *Gitlab {
	result := &Gitlab{
		Server: *NewServer(),
	}
	result.Server.Address = "gitlab.com"
	result.Server.Port = 443
	result.Server.TLS = &TLS{}
	result.GroupFilter.MinAccessLevel = gitlab.MinimalAccessPermissions

	return result
}

func (g *Gitlab) UserAccessControlList() (acls map[string]access.AccessRuler) {
	if len(g.RealmACLs) == 0 {
		// allow anyone into the default realm
		// if nothing has been configured
		acls = map[string]access.AccessRuler{
			"": access.UserDefaultRequirement,
		}

		return
	}

	acls = make(map[string]access.AccessRuler, len(g.RealmACLs))
	for realm, rules := range g.RealmACLs {
		acls[realm] = rules.UserRules()
	}

	return
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
