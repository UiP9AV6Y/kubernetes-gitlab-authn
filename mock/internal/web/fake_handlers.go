package web

import (
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"
)

const (
	HeaderRequestID          = "X-Request-Id"
	HeaderRuntime            = "X-Runtime"
	HeaderRateLimitLimit     = "RateLimit-Limit"
	HeaderRateLimitObserved  = "RateLimit-Observed"
	HeaderRateLimitRemaining = "RateLimit-Remaining"
	HeaderRateLimitReset     = "RateLimit-Reset"
	HeaderRateLimitResetTime = "RateLimit-ResetTime"
)

type FakeRequestIdentificationHandler struct {
	h http.Handler
}

func NewFakeRequestIdentificationHandler(handler http.Handler) *FakeRequestIdentificationHandler {
	result := &FakeRequestIdentificationHandler{
		h: handler,
	}

	return result
}

func (h *FakeRequestIdentificationHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	src := make([]byte, 8)
	now := time.Now().UnixNano()

	binary.NativeEndian.PutUint64(src, uint64(now))

	rid := hex.EncodeToString(src)

	w.Header().Set(HeaderRequestID, rid)

	h.h.ServeHTTP(w, req)
}

type FakeRuntimeHandler struct {
	h http.Handler
	d time.Duration
}

func NewFakeRuntimeHandler(handler http.Handler) *FakeRuntimeHandler {
	result := &FakeRuntimeHandler{
		h: handler,
	}

	return result
}

func (h *FakeRuntimeHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	rt := strconv.FormatFloat(h.d.Seconds(), 'f', -1, 32)
	// fake measurement by using the duration of the last request
	// a proper timing implementation would intercept Write and WriteHeader of w
	// in order to inject the current request duration into the response before
	// the wrapped header actually writes anything to the wire.
	w.Header().Set(HeaderRuntime, rt)

	start := time.Now()
	h.h.ServeHTTP(w, req)
	t := time.Now()
	h.d = t.Sub(start) // store duration for next request
}

type FakeRateLimitHandler struct {
	l int64
	w time.Duration
	h http.Handler
}

func NewFakeRateLimitHandler(limit uint16, window time.Duration, handler http.Handler) *FakeRateLimitHandler {
	if limit < 60 {
		// ensure limit can never be below the faked usage meter
		limit = 60
	}

	result := &FakeRateLimitHandler{
		l: int64(limit),
		w: window,
		h: handler,
	}

	return result
}

func (h *FakeRateLimitHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	reset := time.Now().Add(h.w).UTC()
	usage := int64(reset.Second())
	limit := strconv.FormatInt(h.l, 10)
	observed := strconv.FormatInt(usage, 10)
	remaining := strconv.FormatInt(h.l-usage, 10)

	w.Header().Set(HeaderRateLimitLimit, limit)
	w.Header().Set(HeaderRateLimitObserved, observed)
	w.Header().Set(HeaderRateLimitRemaining, remaining)
	w.Header().Set(HeaderRateLimitReset, strconv.FormatInt(reset.Unix(), 10))
	w.Header().Set(HeaderRateLimitResetTime, reset.Format(http.TimeFormat))

	h.h.ServeHTTP(w, req)
}
