package main

import (
	"net/http"
	"net/http/pprof"

	pyroprof "github.com/grafana/pyroscope-go/godeltaprof/http/pprof"

	slogadapter "github.com/UiP9AV6Y/go-slog-adapter"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
)

func newProfileRouter(_ *slogadapter.SlogAdapter, cfg *config.Profile) (http.Handler, error) {
	router := http.NewServeMux()

	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/"), pprof.Index)
	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/cmdline"), pprof.Cmdline)
	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/profile"), pprof.Profile)
	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/symbol"), pprof.Symbol)
	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/trace"), pprof.Trace)

	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/delta_heap"), pyroprof.Heap)
	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/delta_block"), pyroprof.Block)
	router.HandleFunc(http.MethodGet+" "+cfg.Server.HandlerPath("pprof/delta_mutex"), pyroprof.Mutex)

	return router, nil
}
