package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/log"
)

func newMetricsRouter(reg *prometheus.Registry, logger *log.Adapter, cfg *config.Metrics) (http.Handler, error) {
	router := http.NewServeMux()
	opts := promhttp.HandlerOpts{
		ErrorLog:            logger,
		ErrorHandling:       promhttp.ContinueOnError,
		MaxRequestsInFlight: cfg.RequestLimit,
		Timeout:             cfg.RequestTimeout,
		Registry:            reg,
	}
	handler := promhttp.HandlerFor(reg, opts)

	router.Handle(cfg.Server.HandlerPath(""), handler)

	return router, nil
}
