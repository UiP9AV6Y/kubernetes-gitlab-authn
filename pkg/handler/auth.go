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
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/metrics"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/tracing"
)

var (
	ErrMissingToken   = errors.New("Missing token")
	ErrMalformedToken = errors.New("Token validation failed")
)

const (
	KindTokenReview = "TokenReview"
)

const unauthorizedUsername = "n/a"

type AuthHandler struct {
	client *gitlab.Client
	logger *slog.Logger
	stats  *metrics.Metrics

	preflight func(string) bool

	listGroups *gitlab.ListGroupsOptions
	userInfo   *access.UserInfoOptions

	userAuth  map[string]userauthz.Authorizer
	userCache *cache.UserInfoCache
}

func NewAuthHandler(client *gitlab.Client, logger *slog.Logger, opts ...func(*AuthHandler)) (result *AuthHandler, err error) {
	listGroups := new(gitlab.ListGroupsOptions)
	userInfo := new(access.UserInfoOptions)
	userAuth := map[string]userauthz.Authorizer{
		"": userauthz.AlwaysAllowAuthorizer,
	}
	userCache := cache.NewUserInfoCache(1 * time.Hour)
	preflight := func(_ string) bool {
		return true
	}
	result = &AuthHandler{
		client:     client,
		logger:     logger,
		preflight:  preflight,
		listGroups: listGroups,
		userInfo:   userInfo,
		userAuth:   userAuth,
		userCache:  userCache,
	}

	for _, o := range opts {
		o(result)
	}

	if result.stats == nil {
		result.stats, err = metrics.NewDefault()
	}

	return
}

func WithAuthGroupFilter(v *gitlab.ListGroupsOptions) func(*AuthHandler) {
	return func(h *AuthHandler) {
		h.listGroups = v
	}
}

func WithAuthUserTransform(v *access.UserInfoOptions) func(*AuthHandler) {
	return func(h *AuthHandler) {
		h.userInfo = v
	}
}

func WithAuthUserACLs(v map[string]userauthz.Authorizer) func(*AuthHandler) {
	return func(h *AuthHandler) {
		h.userAuth = v
	}
}

func WithAuthUserCache(v *cache.UserInfoCache) func(*AuthHandler) {
	return func(h *AuthHandler) {
		h.userCache = v
	}
}

func WithAuthMetrics(v *metrics.Metrics) func(*AuthHandler) {
	return func(h *AuthHandler) {
		h.stats = v
	}
}

func WithAuthTokenValidator(v func(string) bool) func(*AuthHandler) {
	return func(h *AuthHandler) {
		h.preflight = v
	}
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	s := r.PathValue("realm")
	t, m, err := parseReviewToken(r.Body)
	if err != nil || !h.preflight(t) {
		if err == nil {
			err = ErrMalformedToken
		}
		h.logger.Info("Invalid authentication request received", "err", err)
		h.stats.AuthMalformed(s)
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
			h.logger.Info("Authentication failed", "user", i.Username, "err", err)
			cache.SetUserInfo(h.userCache, t, i)
			h.stats.AuthNotFound(s)
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
			h.stats.AuthNotFound(s)
			h.rejectReview(w, m, "repeated authentication failure", http.StatusUnauthorized)
			return
		}
	}

	err = h.authorize(r.Context(), s, i)
	if err != nil {
		h.logger.Info("Authorization failed", "user", i.Username, "realm", s, "err", err)
		h.stats.AuthUnauthorized(s)
		h.rejectReview(w, m, "precondition failed", http.StatusUnauthorized)
		return
	}

	h.logger.Info("Authorization accepted", "user", i.Username, "realm", s)
	h.stats.AuthSuccess(s)
	h.acceptReview(w, m, i)
}

func (h *AuthHandler) authenticate(ctx context.Context, token string) (user *gitlab.User, groups []*gitlab.Group, err error) {
	request := tracing.RequestIdentifierFromContext(ctx)
	start := time.Now()
	user, _, err = h.client.Users.CurrentUser(
		gitlab.WithContext(ctx),
		gitlab.WithToken(gitlab.PrivateToken, token),
		gitlab.WithHeader(HeaderRequestId, request),
	)
	h.stats.GitlabRequest("users", time.Since(start))
	if err != nil {
		user = &gitlab.User{
			Username: unauthorizedUsername,
		}
		return
	}

	start = time.Now()
	groups, _, err = h.client.Groups.ListGroups(h.listGroups,
		gitlab.WithContext(ctx),
		gitlab.WithToken(gitlab.PrivateToken, token),
		gitlab.WithHeader(HeaderRequestId, request),
	)
	h.stats.GitlabRequest("groups", time.Since(start))
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

	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(dto)
}

func parseReviewToken(b io.Reader) (token string, header meta.TypeMeta, err error) {
	dto := &authentication.TokenReview{}
	err = json.NewDecoder(b).Decode(dto)
	if err != nil {
		return
	}

	header = dto.TypeMeta
	token = dto.Spec.Token

	if header.APIVersion == "" {
		header.APIVersion = authentication.SchemeGroupVersion.String()
	}

	if header.Kind == "" {
		header.Kind = KindTokenReview
	}

	if token == "" {
		err = ErrMissingToken
	}

	return
}
