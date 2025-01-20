package main

import (
	"net/http"

	gitlab "gitlab.com/gitlab-org/api/client-go"

	slogadapter "github.com/UiP9AV6Y/go-slog-adapter"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/cache"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/handler"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/metrics"
)

func newAppRouter(reg *metrics.Metrics, users *cache.UserInfoCache, logger *slogadapter.SlogAdapter, cfg *config.Config) (http.Handler, error) {
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

	authHandler, err := handler.NewAuthHandler(apiClient, logger.Logger(),
		handler.WithAuthGroupFilter(cfg.Gitlab.GroupFilter.ListOptions()),
		handler.WithAuthTokenValidator(cfg.Gitlab.TokenValidator()),
		handler.WithAuthUserTransform(cfg.Gitlab.UserInfoOptions()),
		handler.WithAuthUserACLs(cfg.Realms.UserAccessControlList()),
		handler.WithAuthUserCache(users),
		handler.WithAuthMetrics(reg),
	)
	if err != nil {
		return nil, err
	}

	webOpts := handler.FilesystemHandlerOpts{
		GitlabURL:   baseURL,
		Description: cfg.Web.Description,
	}
	webHandler, err := handler.FilesystemHandlerFor(cfg.Web.Path, webOpts)
	if err != nil {
		return nil, err
	}

	redirHandler := http.RedirectHandler("/about/", http.StatusSeeOther)

	router.Handle(http.MethodGet+" "+cfg.Server.HandlerPath("{$}"), redirHandler)
	router.Handle(http.MethodGet+" "+cfg.Server.HandlerPath("index.html"), redirHandler)
	router.Handle(http.MethodGet+" "+cfg.Server.HandlerPath("about/"), http.StripPrefix("/about", webHandler))
	router.Handle(http.MethodPost+" "+cfg.Server.HandlerPath("authenticate"), authHandler)
	router.Handle(http.MethodPost+" "+cfg.Server.HandlerPath("authenticate/{realm}"), authHandler)

	return router, nil
}
