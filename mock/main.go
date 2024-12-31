package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	logflags "github.com/UiP9AV6Y/go-slog-adapter/stdflags"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/web"
)

func run(o, e io.Writer, argv ...string) int {
	name := filepath.Base(argv[0])
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	strict := fs.Bool("auth.strict", false, "Use equality instead of substring comparison for tokens")
	listen := fs.String("web.listen-address", ":8080", "Addresses to listen for incoming HTTP requests")
	log := logflags.NewLogFlags(fs)

	if err := fs.Parse(argv[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}

		fmt.Fprintf(e, "%s, try --help\n", err)
		return 1
	}

	user := model.SelectUserByTokenQuery(*strict)
	groups := model.SelectGroupsByTokenQuery(*strict)
	logger := log.Adapter(o, nil).Logger()

	http.HandleFunc("/", web.NotFoundHandler(logger))
	http.HandleFunc("/api/v4/user", web.MeHandler(user, logger))
	http.HandleFunc("/api/v4/groups", web.GroupsHandler(groups, logger))

	logger.Info("Listening on", "address", *listen)
	if err := http.ListenAndServe(*listen, nil); err != nil {
		logger.Error("HTTP Server error", "err", err)
		return 1
	}

	return 0
}

func main() {
	os.Exit(run(os.Stdout, os.Stderr, os.Args...))
}
