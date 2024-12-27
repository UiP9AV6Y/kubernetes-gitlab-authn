package handler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

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
)

type AuthHandlerOpts struct {
	AttributesAsGroups bool

	GroupsOwnedOnly      bool
	GroupsTopLevelOnly   bool
	GroupsMinAccessLevel gitlab.AccessLevelValue
	GroupsFilter         string

	UserACLs access.UserRealmRuler
}

func NewAuthHandlerOpts() *AuthHandlerOpts {
	result := &AuthHandlerOpts{
		GroupsMinAccessLevel: gitlab.MinimalAccessPermissions,
		UserACLs:             access.NewDefaultUserRealmRuler(),
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

	userAuth access.UserRealmRuler
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

func (h *AuthHandler) Authorize(user *gitlab.User, groups []*gitlab.Group, realm string) (err error) {
	err = h.userAuth.AuthorizeUser(realm, user)
	if err != nil {
		return
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
		gids[i] = g.Path
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
		groups = append(groups, "gitlab::2fa")
	}
	if user.Bot {
		groups = append(groups, "gitlab::bot")
	}
	if user.IsAdmin {
		groups = append(groups, "gitlab::admin")
	}
	if user.IsAuditor {
		groups = append(groups, "gitlab::auditor")
	}
	if user.External {
		groups = append(groups, "gitlab::external")
	}

	return groups
}

func userAttributeExtra(user *gitlab.User) map[string]authentication.ExtraValue {
	extra := map[string]authentication.ExtraValue{
		"gitlab-2fa":       boolExtraValue(user.TwoFactorEnabled),
		"gitlab-bot":       boolExtraValue(user.Bot),
		"gitlab-admin":     boolExtraValue(user.IsAdmin),
		"gitlab-auditor":   boolExtraValue(user.IsAuditor),
		"gitlab-external":  boolExtraValue(user.External),
		"gitlab-namespace": intExtraValue(user.NamespaceID),
	}

	for _, attr := range user.CustomAttributes {
		extra[attr.Key] = stringExtraValue(attr.Value)
	}

	return extra
}

func stringExtraValue(s string) authentication.ExtraValue {
	v := []string{s}
	return authentication.ExtraValue(v)
}

func boolExtraValue(b bool) authentication.ExtraValue {
	v := []string{strconv.FormatBool(b)}
	return authentication.ExtraValue(v)
}

func intExtraValue(i int) authentication.ExtraValue {
	v := []string{strconv.FormatInt(int64(i), 10)}
	return authentication.ExtraValue(v)
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
