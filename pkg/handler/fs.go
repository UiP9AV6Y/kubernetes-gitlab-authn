package handler

import (
	"bytes"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

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

func NewFilesystemHandlerOpts() *FilesystemHandlerOpts {
	extraData := map[string]interface{}{}
	result := &FilesystemHandlerOpts{
		Name:      "gitlab-authn",
		Version:   version.Version(),
		GitlabURL: DefaultGitlabURL,
		StartTime: time.Now(),
		ExtraData: extraData,
	}

	return result
}

type FilesystemHandler struct {
	landingPage []byte
	fallback    http.Handler
	modtime     time.Time
}

func NewFilesystemHandler(dir string, opts *FilesystemHandlerOpts) (*FilesystemHandler, error) {
	if opts == nil {
		opts = NewFilesystemHandlerOpts()
	}

	landingPage := filepath.Join(dir, "index.html")
	landingData, err := os.ReadFile(landingPage)
	if err != nil {
		return nil, err
	}

	landingView, err := template.New("/index.html").Parse(string(landingData))
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := landingView.Execute(&buf, opts); err != nil {
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

	if r.Method != http.MethodGet {
		http.Error(w, httpStatusMethodNotAllowed, http.StatusMethodNotAllowed)
		return
	}

	if route == "/index.html" {
		w.Header().Set("Content-Type", contentTypeHTML)
		w.Header().Set("Last-Modified", h.modtime.Format(http.TimeFormat))
		_, _ = w.Write(h.landingPage)
	} else {
		h.fallback.ServeHTTP(w, r)
	}
}
