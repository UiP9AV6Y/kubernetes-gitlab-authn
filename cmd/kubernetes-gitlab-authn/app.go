package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	slogadapter "github.com/UiP9AV6Y/go-slog-adapter"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/cache"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/handler"
)

func newAppRouter(reg *prometheus.Registry, users *cache.UserInfoCache, logger *slogadapter.SlogAdapter, cfg *config.Config) (http.Handler, error) {
	router := http.NewServeMux()
	baseURL, err := cfg.Gitlab.URL()
	if err != nil {
		return nil, err
	}

	httpClient, err := cfg.Gitlab.HTTPClient()
	if err != nil {
		return nil, err
	}

	apiClient, err := gitlab.NewClient("",
		gitlab.WithBaseURL(baseURL.String()),
		gitlab.WithHTTPClient(httpClient),
		gitlab.WithCustomLeveledLogger(logger.Logger()),
	)
	if err != nil {
		return nil, err
	}

	authOpts := &handler.AuthHandlerOpts{
		AttributesAsGroups:   cfg.Gitlab.AttributesAsGroups,
		InactivityTimeout:    cfg.Gitlab.InactivityTimeout,
		GroupsOwnedOnly:      cfg.Gitlab.GroupFilter.OwnedOnly,
		GroupsTopLevelOnly:   cfg.Gitlab.GroupFilter.TopLevelOnly,
		GroupsMinAccessLevel: cfg.Gitlab.GroupFilter.MinAccessLevel,
		GroupsFilter:         cfg.Gitlab.GroupFilter.Name,
		GroupsLimit:          int(cfg.Gitlab.GroupFilter.Limit),
		UserACLs:             cfg.Realms.UserAccessControlList(),
		UserCache:            users,
	}
	authHandler, err := handler.NewAuthHandler(apiClient, logger.Logger(), authOpts)
	if err != nil {
		return nil, err
	}

	webOpts := handler.NewFilesystemHandlerOpts()
	webOpts.GitlabURL = baseURL
	webOpts.Description = cfg.Web.Description
	webHandler, err := handler.NewFilesystemHandler(cfg.Web.Path, webOpts)
	if err != nil {
		return nil, err
	}

	router.Handle(cfg.Server.HandlerPath(""), webHandler)
	router.Handle(cfg.Server.HandlerPath("authenticate"), authHandler)
	router.Handle(cfg.Server.HandlerPath("authenticate/{realm}"), authHandler)

	return router, nil
}
