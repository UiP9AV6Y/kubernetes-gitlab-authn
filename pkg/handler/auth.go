package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	authentication "k8s.io/api/authentication/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/access"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/log"
)

var (
	ErrUserLocked      = errors.New("User is locked")
	ErrUserUnconfirmed = errors.New("User has not yet confirmed their account")
	ErrUserRobot       = errors.New("Robot access has been denied")
	ErrUser2FA         = errors.New("User does not have 2FA set up")
	ErrMissingToken    = errors.New("Missing token")
)

const (
	KindTokenReview                 = "TokenReview"
	DefaultAuthenticationAPIVersion = "authentication.k8s.io/v1"

	// GitlabKeyNamespace is the key namespace used in a user's "extra"
	// to represent the various Gitlab specific account attributes
	GitlabKeyNamespace = "gitlab-authn.kubernetes.io/"
	// GitlabAttributesKey is the key used in a user's "extra" to specify
	// the Gitlab specific account attributes
	GitlabAttributesKey = GitlabKeyNamespace + "user-attributes"
)

type AuthHandlerOpts struct {
	AttributesAsGroups bool

	GroupsOwnedOnly      bool
	GroupsTopLevelOnly   bool
	GroupsMinAccessLevel gitlab.AccessLevelValue
	GroupsFilter         string

	UserACLs map[string]access.AccessRuler
}

func NewAuthHandlerOpts() *AuthHandlerOpts {
	acls := map[string]access.AccessRuler{
		"": access.UserDefaultRequirement,
	}
	result := &AuthHandlerOpts{
		GroupsMinAccessLevel: gitlab.MinimalAccessPermissions,
		UserACLs:             acls,
	}

	return result
}

func (o *AuthHandlerOpts) ListGroupsOptions() *gitlab.ListGroupsOptions {
	result := &gitlab.ListGroupsOptions{
		MinAccessLevel: &o.GroupsMinAccessLevel,
	}

	if o.GroupsFilter != "" {
		result.Search = &o.GroupsFilter
	}

	if o.GroupsOwnedOnly {
		result.Owned = &o.GroupsOwnedOnly
	}

	if o.GroupsTopLevelOnly {
		result.TopLevelOnly = &o.GroupsTopLevelOnly
	}

	return result
}

type AuthHandler struct {
	client     *gitlab.Client
	logger     *log.Adapter
	listGroups *gitlab.ListGroupsOptions
	attrGroups bool

	userAuth map[string]access.AccessRuler
}

func NewAuthHandler(client *gitlab.Client, logger *log.Adapter, opts *AuthHandlerOpts) (*AuthHandler, error) {
	if opts == nil {
		opts = NewAuthHandlerOpts()
	}

	result := &AuthHandler{
		client:     client,
		logger:     logger,
		listGroups: opts.ListGroupsOptions(),
		attrGroups: opts.AttributesAsGroups,
		userAuth:   opts.UserACLs,
	}

	return result, nil
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	t, m, err := parseReviewToken(r.Body)
	if err != nil {
		h.logger.Info("Invalid authentication request received", "err", err)
		w.WriteHeader(http.StatusBadRequest)
		h.rejectReview(w, m, "malformed review request")
		return
	}

	u, g, err := h.Authenticate(r.Context(), t)
	if err != nil {
		h.logger.Info("Authentication failed", "user", u.Username, "err", err)
		w.WriteHeader(http.StatusUnauthorized)
		h.rejectReview(w, m, "unable to review request")
		return
	}

	s := r.PathValue("realm")
	err = h.Authorize(u, g, s)
	if err != nil {
		h.logger.Info("Authorization failed", "user", u.Username, "realm", s, "err", err)
		w.WriteHeader(http.StatusUnauthorized)
		h.rejectReview(w, m, "precondition failed")
		return
	}

	i := h.UserInfo(u, g)
	h.logger.Info("Authentication accepted", "user", u.Username)
	w.WriteHeader(http.StatusOK)
	h.acceptReview(w, m, i)
}

func (h *AuthHandler) Authenticate(ctx context.Context, token string) (user *gitlab.User, groups []*gitlab.Group, err error) {
	user, _, err = h.client.Users.CurrentUser(
		gitlab.WithContext(ctx),
		gitlab.WithToken(gitlab.PrivateToken, token),
	)
	if err != nil {
		user = &gitlab.User{
			Username: "n/a",
		}
		return
	}

	groups, _, err = h.client.Groups.ListGroups(h.listGroups,
		gitlab.WithContext(ctx),
		gitlab.WithToken(gitlab.PrivateToken, token),
	)
	if err != nil {
		return
	}

	return
}

func (h *AuthHandler) Authorize(user *gitlab.User, groups []*gitlab.Group, realm string) error {
	userAuth, ok := h.userAuth[realm]
	if !ok {
		return fmt.Errorf("No such authentication realm %q", realm)
	}

	ok = userAuth.Authorize(user, groups)
	if !ok {
		return fmt.Errorf("user %q is not authorized to access realm %q", user.Username, realm)
	}

	return nil
}

func (h *AuthHandler) UserInfo(user *gitlab.User, groups []*gitlab.Group) authentication.UserInfo {
	var gids []string
	if h.attrGroups {
		agids := userAttributeGroups(user)
		gids = make([]string, len(groups), len(groups)+len(agids))
		gids = append(gids, agids...)
	} else {
		gids = make([]string, len(groups))
	}

	for i, g := range groups {
		gids[i] = strings.ReplaceAll(g.FullPath, "/", ":")
	}

	extra := userAttributeExtra(user)
	info := authentication.UserInfo{
		Username: user.Username,
		UID:      strconv.FormatInt(int64(user.ID), 10),
		Groups:   gids,
		Extra:    extra,
	}

	return info
}

func userAttributeGroups(user *gitlab.User) []string {
	groups := make([]string, 0, 5)

	if user.TwoFactorEnabled {
		groups = append(groups, "gitlab:2fa")
	}
	if user.Bot {
		groups = append(groups, "gitlab:bot")
	}
	if user.IsAdmin {
		groups = append(groups, "gitlab:admin")
	}
	if user.IsAuditor {
		groups = append(groups, "gitlab:auditor")
	}
	if user.External {
		groups = append(groups, "gitlab:external")
	}

	return groups
}

func userAttributeExtra(user *gitlab.User) map[string]authentication.ExtraValue {
	attrs := make([]string, 0, 5)
	if user.TwoFactorEnabled {
		attrs = append(attrs, "2fa")
	}
	if user.Bot {
		attrs = append(attrs, "bot")
	}
	if user.IsAdmin {
		attrs = append(attrs, "admin")
	}
	if user.IsAuditor {
		attrs = append(attrs, "auditor")
	}
	if user.External {
		attrs = append(attrs, "external")
	}

	extra := map[string]authentication.ExtraValue{
		GitlabAttributesKey: attrs,
	}

	for _, attr := range user.CustomAttributes {
		extra[GitlabKeyNamespace+attr.Key] = []string{attr.Value}
	}

	return extra
}

func (h *AuthHandler) rejectReview(w http.ResponseWriter, header meta.TypeMeta, err string) {
	status := authentication.TokenReviewStatus{
		Error: err,
	}

	writeReview(w, header, status)
}

func (h *AuthHandler) acceptReview(w http.ResponseWriter, header meta.TypeMeta, info authentication.UserInfo) {
	status := authentication.TokenReviewStatus{
		Authenticated: true,
		User:          info,
	}

	writeReview(w, header, status)
}

func writeReview(w http.ResponseWriter, header meta.TypeMeta, status authentication.TokenReviewStatus) {
	dto := &authentication.TokenReview{
		ObjectMeta: meta.ObjectMeta{CreationTimestamp: meta.Now()},
		TypeMeta:   header,
		Status:     status,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(dto)
}

func parseReviewToken(b io.Reader) (token string, header meta.TypeMeta, err error) {
	header = meta.TypeMeta{APIVersion: DefaultAuthenticationAPIVersion, Kind: KindTokenReview}
	dto := &authentication.TokenReview{}
	err = json.NewDecoder(b).Decode(dto)
	if err != nil {
		return
	}

	header = dto.TypeMeta
	token = dto.Spec.Token

	if header.APIVersion == "" {
		header.APIVersion = DefaultAuthenticationAPIVersion
	}

	if header.Kind == "" {
		header.Kind = KindTokenReview
	}

	if token == "" {
		err = ErrMissingToken
	}

	return
}
