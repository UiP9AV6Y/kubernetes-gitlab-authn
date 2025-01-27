package handler

import (
	"fmt"
	"net/http"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/health"
)

type HealthHandlerOpts struct {
	// Header value for the reponses
	ContentType string
	// Response body for healthy states
	HealthyResponse string
	// Response body for unhealthy states
	UnhealthyResponse string
}

type HealthHandler struct {
	status *health.Health
	ok     func(http.ResponseWriter)
	fail   func(http.ResponseWriter)
}

func HealthHandlerFor(status *health.Health, opts HealthHandlerOpts) (*HealthHandler, error) {
	var ct, healthy, unhealthy string

	switch opts.ContentType {
	case "", ContentTypeText:
		ct = ContentTypeText
		healthy = "OK"
		unhealthy = "FAIL"
	case ContentTypeHTML:
		ct = ContentTypeHTML
		healthy = `<h1 style="color: green">OK</h1>`
		unhealthy = `<h1 style="color: red">FAIL</h1>`
	case ContentTypeJSON:
		ct = ContentTypeJSON
		healthy = `{"status": "OK"}`
		unhealthy = `{"status": "FAIL"}`
	}

	if opts.HealthyResponse != "" {
		healthy = opts.HealthyResponse
	}

	if opts.UnhealthyResponse != "" {
		healthy = opts.UnhealthyResponse
	}

	ok := func(w http.ResponseWriter) {
		w.Header().Set(HeaderContentType, ct)
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, healthy)
	}
	fail := func(w http.ResponseWriter) {
		w.Header().Set(HeaderContentType, ct)
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, unhealthy)
	}
	result := &HealthHandler{
		ok:     ok,
		fail:   fail,
		status: status,
	}

	return result, nil
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(HeaderContentTypeOptions, "nosniff")

	if h.status.Status() {
		h.ok(w)
	} else {
		h.fail(w)
	}
}
