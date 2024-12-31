package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"

	bicol "github.com/UiP9AV6Y/buildinfo/prometheus/collector"
	slogadapter "github.com/UiP9AV6Y/go-slog-adapter"

	"golang.org/x/sync/errgroup"

	logflags "github.com/UiP9AV6Y/go-slog-adapter/stdflags"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
	cfgflags "github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config/stdflags"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/metrics"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/version"
)

const envPrefix = "GITLAB_AUTHN_"

func newHTTPServer(h http.Handler, ctx context.Context) *http.Server {
	result := &http.Server{
		Handler: h,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	return result
}

func runServers(name string, config *config.Config, logger *slogadapter.SlogAdapter) (err error) {
	var router http.Handler
	var server *http.Server

	mainCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	registry := prometheus.NewRegistry()
	tasks, tasksCtx := errgroup.WithContext(mainCtx)
	servers := &serverManager{
		logger: logger.Logger(),
		ctx:    tasksCtx,
	}

	buildinfo := bicol.New(version.BuildInfo(), metrics.Namespace)
	err = registry.Register(buildinfo)
	if err != nil {
		return
	}

	router, err = newAppRouter(registry, logger, config)
	if err != nil {
		return err
	}

	server = newHTTPServer(router, mainCtx)
	bootup, shutdown := servers.Task("app", server, config.Server)
	queue := []serverTask{bootup, shutdown}

	if config.Metrics.Port > 0 {
		router, err = newMetricsRouter(registry, logger, config.Metrics)
		if err != nil {
			return err
		}

		server = newHTTPServer(router, mainCtx)
		bootup, shutdown = servers.Task("metrics", server, &config.Metrics.Server)
		queue = append(queue, bootup, shutdown)
	}

	if config.Health.Port > 0 {
		router, err = newHealthRouter(logger, config.Health)
		if err != nil {
			return err
		}

		server = newHTTPServer(router, mainCtx)
		bootup, shutdown = servers.Task("health", server, &config.Health.Server)
		queue = append(queue, bootup, shutdown)
	}

	logger.Info("Starting "+name, "version", version.Version())
	for _, t := range queue {
		tasks.Go(t)
	}

	if err := tasks.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		return err
	}

	return nil
}

func run(o, e io.Writer, argv ...string) int {
	name := filepath.Base(argv[0])
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	ver := fs.Bool("version", false, "Print the version number and exit")
	log := logflags.NewEnvLogFlags(fs, envPrefix)
	cfg := cfgflags.NewEnvConfigFlags(fs, envPrefix)

	if err := log.ParseEnv(); err != nil {
		fmt.Fprintf(e, "%s, try --help\n", err)
		return 1
	}

	if err := cfg.ParseEnv(); err != nil {
		fmt.Fprintf(e, "%s, try --help\n", err)
		return 1
	}

	if err := fs.Parse(argv[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		fmt.Fprintf(e, "%s, try --help\n", err)
		return 1
	}

	if *ver {
		fmt.Fprintln(o, version.Version())
		return 0
	}

	adapter := log.Adapter(o, nil)
	settings := cfg.Config()

	if err := runServers(name, settings, adapter); err != nil {
		adapter.Logger().Error("Application terminated", "err", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args...))
}
