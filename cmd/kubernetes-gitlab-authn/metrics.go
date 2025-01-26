package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	slogadapter "github.com/UiP9AV6Y/go-slog-adapter"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
)

func newMetricsRouter(reg *prometheus.Registry, logger *slogadapter.SlogAdapter, cfg *config.Metrics) (http.Handler, error) {
	router := http.NewServeMux()
	opts := promhttp.HandlerOpts{
		ErrorLog:            logger,
		ErrorHandling:       promhttp.ContinueOnError,
		MaxRequestsInFlight: cfg.RequestLimit,
		Timeout:             cfg.RequestTimeout,
		Registry:            reg,
	}
	handler := promhttp.HandlerFor(reg, opts)

	router.Handle(http.MethodGet+" "+cfg.Server.HandlerPath(""), handler)

	return router, nil
}
