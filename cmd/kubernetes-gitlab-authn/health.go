package main

import (
	"fmt"
	"net/http"

	slogadapter "github.com/UiP9AV6Y/go-slog-adapter"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
)

func newHealthRouter(_ *slogadapter.SlogAdapter, cfg *config.Health) (http.Handler, error) {
	router := http.NewServeMux()
	handler := func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "OK")
	}

	router.HandleFunc(cfg.Server.HandlerPath("health"), handler)
	router.HandleFunc(cfg.Server.HandlerPath("ready"), handler)

	return router, nil
}
