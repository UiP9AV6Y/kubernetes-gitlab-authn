package main

import (
	"net/http"

	slogadapter "github.com/UiP9AV6Y/go-slog-adapter"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/handler"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/health"
)

func newHealthRouter(status *health.Health, _ *slogadapter.SlogAdapter, cfg *config.Health) (http.Handler, error) {
	router := http.NewServeMux()
	opts := handler.HealthHandlerOpts{}
	handler, err := handler.HealthHandlerFor(status, opts)
	if err != nil {
		return nil, err
	}

	router.Handle(http.MethodGet+" "+cfg.Server.HandlerPath("health"), handler)
	router.Handle(http.MethodGet+" "+cfg.Server.HandlerPath("ready"), handler)

	return router, nil
}
