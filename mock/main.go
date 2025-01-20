package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	logflags "github.com/UiP9AV6Y/go-slog-adapter/stdflags"

	dao "github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
	mdao "github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model/memory"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/web"
)

func run(o, e io.Writer, argv ...string) int {
	name := filepath.Base(argv[0])
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	groups := fs.Uint64("mock.feature-groups", 50, "Number of mock feature groups to create")
	prefix := fs.String("mock.token-prefix", "glpat-", "Prefix to use for generated authentication tokens")
	listen := fs.String("web.listen-address", ":8080", "Addresses to listen for incoming HTTP requests")
	rtTime := fs.Duration("rate-limit.interval", 1*time.Minute, "Fake rate limit interval to report to clients")
	rtSize := fs.Int64("rate-limit.quota", 100, "Fake rate limit quota to report to clients")
	log := logflags.NewLogFlags(fs)

	if err := fs.Parse(argv[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		fmt.Fprintf(e, "%s, try --help\n", err)
		return 1
	}

	logger := log.Adapter(o, nil).Logger()

	data, err := mdao.NewDataAccess()
	if err != nil {
		logger.Error("Data access setup failed", "err", err)
		return 1
	}

	mocks := &dao.Mocks{
		TokenPrefix: *prefix,
		GroupCount:  *groups,
	}
	if err := mocks.Create(data); err != nil {
		logger.Error("Mock seeding failed", "err", err)
		return 1
	}

	router := http.NewServeMux()
	router.Handle("/", web.NotFoundHandler(logger))
	router.Handle("/api/v4/user", web.MeHandler(data, logger))
	router.Handle("/api/v4/groups", web.GroupsHandler(data, logger))
	router.Handle("/api/v4/version", web.VersionHandler(logger))
	router.Handle("/api/v4/metadata", web.MetaDataHandler(logger))

	handler := web.NewFakeRuntimeHandler(
		web.NewFakeRequestIdentificationHandler(
			web.NewFakeRateLimitHandler(uint16(*rtSize), *rtTime, router),
		),
	)

	logger.Info("Listening on", "address", *listen)
	if err := http.ListenAndServe(*listen, handler); err != nil {
		logger.Error("HTTP Server error", "err", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args...))
}
