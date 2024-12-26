package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/handler"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/log"
)

func newAppRouter(reg *prometheus.Registry, logger *log.Adapter, cfg *config.Config) (http.Handler, error) {
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
		gitlab.WithCustomLeveledLogger(logger),
	)
	if err != nil {
		return nil, err
	}

	authOpts := &handler.AuthHandlerOpts{
		Require2FA:           cfg.Gitlab.UserFilter.Require2FA,
		RejectBots:           cfg.Gitlab.UserFilter.RejectBots,
		AttributesAsGroups:   cfg.Gitlab.AttributesAsGroups,
		GroupsOwnedOnly:      cfg.Gitlab.GroupFilter.OwnedOnly,
		GroupsTopLevelOnly:   cfg.Gitlab.GroupFilter.TopLevelOnly,
		GroupsMinAccessLevel: cfg.Gitlab.GroupFilter.MinAccessLevel,
		GroupsFilter:         cfg.Gitlab.GroupFilter.Name,
	}
	authHandler, err := handler.NewAuthHandler(apiClient, logger, authOpts)
	if err != nil {
		return nil, err
	}

	webOpts := handler.NewFilesystemHandlerOpts()
	webOpts.GitlabURL = baseURL
	webOpts.Description = cfg.Web.Description
	webHandler, err := handler.NewFilesystemHandler(cfg.Web.Path, logger, webOpts)
	if err != nil {
		return nil, err
	}

	router.Handle(cfg.Server.HandlerPath(""), webHandler)
	router.Handle(cfg.Server.HandlerPath("authenticate"), authHandler)

	return router, nil
}
