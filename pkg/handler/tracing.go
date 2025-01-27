package handler

import (
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/tracing"
)

// TimeRequestIdentifier returns a middleware implementation which
// decorates incoming requests with [HeaderRequestId] headers derived
// from the provided time generator. The algorithm input is using the
// value from Time.Unix as calculation base.
func TimeRequestIdentifier(now func() time.Time, h http.Handler) http.Handler {
	cb := func() uint64 {
		return uint64(now().Unix())
	}
	return RequestIdentifier(cb, h)
}

// SystemRequestIdentifier returns a middleware implementation which
// decorates incoming requests with [HeaderRequestId] headers derived
// from the system time.
func SystemRequestIdentifier(h http.Handler) http.Handler {
	return TimeRequestIdentifier(time.Now, h)
}

// LinearRequestIdentifier returns a middleware implementation which
// decorates incoming requests with [HeaderRequestId] headers using
// incremental values. The step size is defined by inc.
func LinearRequestIdentifier(inc uint64, h http.Handler) http.Handler {
	cb := func() uint64 {
		return inc
	}
	return RequestIdentifier(cb, h)
}

// RequestIdentifier wraps the given handler with a header injector
// using the provided increment provider as request identification mutator/generator.
// The request identifier is injected only if no existing value is present.
func RequestIdentifier(inc func() uint64, h http.Handler) http.Handler {
	counter := new(atomic.Uint64)
	counter.Add(inc()) // prime for use to avoid starting with zero

	middleware := func(w http.ResponseWriter, r *http.Request) {
		rid := r.Header.Get(HeaderRequestId)
		if rid == "" {
			nun := counter.Add(inc())
			rid := strconv.FormatUint(nun, 10)
			r.Header.Set(HeaderRequestId, rid)
		}

		ctx := tracing.NewContextWithRequestIdentifier(r.Context(), rid)
		h.ServeHTTP(w, r.WithContext(ctx))
	}

	return http.HandlerFunc(middleware)
}
