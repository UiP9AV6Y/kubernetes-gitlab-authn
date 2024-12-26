package main

import (
	"fmt"
	"net/http"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/log"
)

func newHealthRouter(_ *log.Adapter, cfg *config.Health) (http.Handler, error) {
	router := http.NewServeMux()
	handler := func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintf(w, "OK")
	}

	router.HandleFunc(cfg.Server.HandlerPath("health"), handler)
	router.HandleFunc(cfg.Server.HandlerPath("ready"), handler)

	return router, nil
}
