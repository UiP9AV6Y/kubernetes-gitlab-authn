package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	authentication "k8s.io/api/authentication/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

	userauthz "github.com/UiP9AV6Y/go-k8s-user-authz"
	"github.com/UiP9AV6Y/go-k8s-user-authz/userinfo"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/access"
)

var (
	ErrMissingToken = errors.New("Missing token")
)

const (
	KindTokenReview                 = "TokenReview"
	DefaultAuthenticationAPIVersion = "authentication.k8s.io/v1"
)

type AuthHandlerOpts struct {
	AttributesAsGroups bool
	InactivityTimeout  time.Duration

	GroupsOwnedOnly      bool
	GroupsTopLevelOnly   bool
	GroupsMinAccessLevel gitlab.AccessLevelValue
	GroupsFilter         string

	UserACLs map[string]userauthz.Authorizer
}

func NewAuthHandlerOpts() *AuthHandlerOpts {
	acls := map[string]userauthz.Authorizer{
		"": userauthz.AlwaysAllowAuthorizer,
	}
	result := &AuthHandlerOpts{
		GroupsMinAccessLevel: gitlab.MinimalAccessPermissions,
		UserACLs:             acls,
	}

	return result
}

func (o *AuthHandlerOpts) UserInfoOptions() *access.UserInfoOptions {
	result := &access.UserInfoOptions{
		AttributesAsGroups: o.AttributesAsGroups,
		DormantTimeout:     o.InactivityTimeout,
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
	logger     *slog.Logger
	listGroups *gitlab.ListGroupsOptions
	userInfo   *access.UserInfoOptions

	userAuth map[string]userauthz.Authorizer
}

func NewAuthHandler(client *gitlab.Client, logger *slog.Logger, opts *AuthHandlerOpts) (*AuthHandler, error) {
	if opts == nil {
		opts = NewAuthHandlerOpts()
	}

	result := &AuthHandler{
		client:     client,
		logger:     logger,
		listGroups: opts.ListGroupsOptions(),
		userInfo:   opts.UserInfoOptions(),
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

	u, g, err := h.authenticate(r.Context(), t)
	if err != nil {
		h.logger.Info("Authentication failed", "user", u.Username, "err", err)
		w.WriteHeader(http.StatusUnauthorized)
		h.rejectReview(w, m, "unable to review request")
		return
	}

	s := r.PathValue("realm")
	i := access.UserInfo(u, g, *h.userInfo)
	err = h.authorize(r.Context(), s, i)
	if err != nil {
		h.logger.Info("Authorization failed", "user", u.Username, "realm", s, "err", err)
		w.WriteHeader(http.StatusUnauthorized)
		h.rejectReview(w, m, "precondition failed")
		return
	}

	h.logger.Info("Authentication accepted", "user", u.Username, "realm", s)
	w.WriteHeader(http.StatusOK)
	h.acceptReview(w, m, i)
}

func (h *AuthHandler) authenticate(ctx context.Context, token string) (user *gitlab.User, groups []*gitlab.Group, err error) {
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

func (h *AuthHandler) authorize(ctx context.Context, realm string, user authentication.UserInfo) error {
	userAuth, ok := h.userAuth[realm]
	if !ok {
		return fmt.Errorf("No such authentication realm %q", realm)
	}

	info := userinfo.NewV1UserInfo(user)
	decision := userAuth.Authorize(ctx, info)
	if decision != userauthz.DecisionAllow {
		return fmt.Errorf("user %q is not authorized to access realm %q", user.Username, realm)
	}

	return nil
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
