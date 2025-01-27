package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/cache"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/config"
)

type serverTaskCallback struct {
	Start func()
	Stop  func()
}

type serverTask func() error

type serverManager struct {
	logger *slog.Logger
	ctx    context.Context
}

func (m *serverManager) CacheTask(cache *cache.UserInfoCache, cb *serverTaskCallback) (serverTask, serverTask) {
	start := func() error {
		m.logger.Info("Starting cache eviction scheduler")
		if cb != nil && cb.Start != nil {
			cb.Start()
		}
		cache.Start()

		return nil
	}
	stop := func() error {
		<-m.ctx.Done()

		m.logger.Info("Cache eviction is terminating")
		if cb != nil && cb.Stop != nil {
			cb.Stop()
		}
		cache.Stop()

		return nil
	}

	return serverTask(start), serverTask(stop)
}

func (m *serverManager) HTTPTask(name string, server *http.Server, config *config.Server, cb *serverTaskCallback) (serverTask, serverTask) {
	start := func() error {
		var ln net.Listener
		var err error

		ln, err = net.Listen("tcp", config.Addr())
		if err != nil {
			return err
		}

		m.logger.Info("Listening on", "address", ln.Addr().String(), "listener", name)
		if cb != nil && cb.Start != nil {
			cb.Start()
		}
		if config.TLS != nil && config.TLS.CertFile != "" && config.TLS.KeyFile != "" {
			m.logger.Info("TLS is enabled.", "listener", name)
			err = server.ServeTLS(ln, config.TLS.CertFile, config.TLS.KeyFile)
		} else {
			m.logger.Info("TLS is disabled.", "listener", name)
			err = server.Serve(ln)
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	}
	stop := func() error {
		<-m.ctx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		m.logger.Info("Server is shutting down", "listener", name)
		if cb != nil && cb.Stop != nil {
			cb.Stop()
		}
		return server.Shutdown(ctx)
	}

	return serverTask(start), serverTask(stop)
}
