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
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/cache"
)

var (
	ErrMissingToken = errors.New("Missing token")
)

const (
	KindTokenReview                 = "TokenReview"
	DefaultAuthenticationAPIVersion = "authentication.k8s.io/v1"
)

const unauthorizedUsername = "n/a"

type AuthHandlerOpts struct {
	AttributesAsGroups bool
	InactivityTimeout  time.Duration

	GroupsOwnedOnly      bool
	GroupsTopLevelOnly   bool
	GroupsMinAccessLevel gitlab.AccessLevelValue
	GroupsFilter         string
	GroupsLimit          int

	UserACLs  map[string]userauthz.Authorizer
	UserCache *cache.UserInfoCache
}

func NewAuthHandlerOpts() *AuthHandlerOpts {
	acls := map[string]userauthz.Authorizer{
		"": userauthz.AlwaysAllowAuthorizer,
	}
	users := cache.NewUserInfoCache(1 * time.Hour)
	result := &AuthHandlerOpts{
		GroupsMinAccessLevel: gitlab.MinimalAccessPermissions,
		UserACLs:             acls,
		UserCache:            users,
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
	list := gitlab.ListOptions{
		Page: 1,
	}
	result := &gitlab.ListGroupsOptions{
		ListOptions: list,
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

	if o.GroupsLimit > 0 {
		result.ListOptions.PerPage = o.GroupsLimit
	}

	if o.GroupsMinAccessLevel > gitlab.MinimalAccessPermissions {
		result.MinAccessLevel = &o.GroupsMinAccessLevel
	}

	return result
}

type AuthHandler struct {
	client     *gitlab.Client
	logger     *slog.Logger
	listGroups *gitlab.ListGroupsOptions
	userInfo   *access.UserInfoOptions

	userAuth  map[string]userauthz.Authorizer
	userCache *cache.UserInfoCache
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
		userCache:  opts.UserCache,
	}

	return result, nil
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if r.Method != http.MethodPost {
		http.Error(w, httpStatusMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	t, m, err := parseReviewToken(r.Body)
	if err != nil {
		h.logger.Info("Invalid authentication request received", "err", err)
		h.rejectReview(w, m, "malformed review request", http.StatusBadRequest)
		return
	}

	var i authentication.UserInfo
	cached := h.userCache.Get(t)
	if cached == nil {
		u, g, err := h.authenticate(r.Context(), t)
		if err != nil {
			i.Username = u.Username      // for logging purposes later on
			i.UID = unauthorizedUsername // mark as invalid
			cache.SetUserInfo(h.userCache, t, i)
			h.logger.Info("Authentication failed", "user", i.Username, "err", err)
			h.rejectReview(w, m, "unable to review request", http.StatusUnauthorized)
			return
		}

		i = access.UserInfo(u, g, *h.userInfo)
		cache.SetUserInfo(h.userCache, t, i)
		h.logger.Debug("Authentication succeeded", "user", i.Username)
	} else {
		i = cached.Value()
		h.logger.Debug("Using cached authentication", "user", i.Username)
		if i.UID == unauthorizedUsername { // previous rejection
			h.logger.Info("Cached authentication failure", "user", i.Username)
			h.rejectReview(w, m, "repeated authentication failure", http.StatusUnauthorized)
			return
		}
	}

	s := r.PathValue("realm")
	err = h.authorize(r.Context(), s, i)
	if err != nil {
		h.logger.Info("Authorization failed", "user", i.Username, "realm", s, "err", err)
		h.rejectReview(w, m, "precondition failed", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Authorization accepted", "user", i.Username, "realm", s)
	h.acceptReview(w, m, i)
}

func (h *AuthHandler) authenticate(ctx context.Context, token string) (user *gitlab.User, groups []*gitlab.Group, err error) {
	user, _, err = h.client.Users.CurrentUser(
		gitlab.WithContext(ctx),
		gitlab.WithToken(gitlab.PrivateToken, token),
	)
	if err != nil {
		user = &gitlab.User{
			Username: unauthorizedUsername,
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

func (h *AuthHandler) rejectReview(w http.ResponseWriter, header meta.TypeMeta, err string, statusCode int) {
	status := authentication.TokenReviewStatus{
		Error: err,
	}

	writeReview(w, header, status, statusCode)
}

func (h *AuthHandler) acceptReview(w http.ResponseWriter, header meta.TypeMeta, info authentication.UserInfo) {
	status := authentication.TokenReviewStatus{
		Authenticated: true,
		User:          info,
	}

	writeReview(w, header, status, http.StatusOK)
}

func writeReview(w http.ResponseWriter, header meta.TypeMeta, status authentication.TokenReviewStatus, statusCode int) {
	dto := &authentication.TokenReview{
		ObjectMeta: meta.ObjectMeta{CreationTimestamp: meta.Now()},
		TypeMeta:   header,
		Status:     status,
	}

	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
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
