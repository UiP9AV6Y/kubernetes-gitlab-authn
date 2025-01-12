package handler

import (
	"bytes"
	htmltemplate "html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/template"
	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/version"
)

var DefaultGitlabURL, _ = url.Parse("https://gitlab.com/")

type FilesystemHandlerOpts struct {
	// Application name
	Name string
	// Version information
	Version string
	// Freeform text describing the purpose of the application
	Description string
	// URL for users to visit for token generation
	GitlabURL *url.URL
	// Process start time to use as cache buster or information
	StartTime time.Time
	// Custom information to be used in the template
	ExtraData map[string]interface{}
}

type FilesystemHandler struct {
	landingPage []byte
	fallback    http.Handler
	modtime     time.Time
}

func FilesystemHandlerFor(dir string, opts FilesystemHandlerOpts) (*FilesystemHandler, error) {
	if opts.Name == "" {
		opts.Name = "gitlab-authn"
	}
	if opts.Version == "" {
		opts.Version = version.Version()
	}
	if opts.GitlabURL == nil {
		opts.GitlabURL = DefaultGitlabURL
	}
	if opts.StartTime.IsZero() {
		opts.StartTime = time.Now()
	}
	if opts.ExtraData == nil {
		opts.ExtraData = map[string]interface{}{}
	}

	landingPage := filepath.Join(dir, "index.html")
	landingData, err := os.ReadFile(landingPage)
	if err != nil {
		return nil, err
	}

	landingView, err := htmltemplate.New("/index.html").Funcs(template.Functions).Parse(string(landingData))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := landingView.Execute(&buf, &opts); err != nil {
		return nil, err
	}

	filesystem := http.FileServer(http.Dir(dir))
	result := &FilesystemHandler{
		landingPage: buf.Bytes(),
		fallback:    filesystem,
		modtime:     opts.StartTime.UTC(),
	}

	return result, nil
}

func (h *FilesystemHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := r.URL.Path
	if route == "/" {
		route = "/index.html"
	}

	if route == "/index.html" {
		w.Header().Set("Content-Type", contentTypeHTML)
		w.Header().Set("Last-Modified", h.modtime.Format(http.TimeFormat))
		_, _ = w.Write(h.landingPage)
	} else {
		h.fallback.ServeHTTP(w, r)
	}
}
