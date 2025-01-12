package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	optsAuthFailures = prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: "authentication",
		Name:      "failures_total",
		Help:      "Number of authentication failures.",
	}
	optsAuthAttempts = prometheus.CounterOpts{
		Namespace: Namespace,
		Subsystem: "authentication",
		Name:      "attempts_total",
		Help:      "Number of authentication attempts.",
	}
	optsGitlabDuration = prometheus.HistogramOpts{
		Namespace: Namespace,
		Subsystem: "gitlab",
		Name:      "request_duration_seconds",
		Help:      "Elapsed time in seconds for HTTP request against Gitlab.",
		Buckets:   []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}
)

const (
	labelRealm   = "realm"
	labelCause   = "cause"
	labelService = "service"
)

const (
	authCauseMalformed    = "malformed"
	authCauseNotFound     = "not_found"
	authCauseUnauthorized = "unauthorized"
)

// Metrics is an abstraction over several measurement trackers.
// Instead of exposing the various counters, gauges, histograms, ...
// this implementation exposes a simplified API for application
// specific scenarios.
type Metrics struct {
	authFailures   *prometheus.CounterVec
	authAttempts   *prometheus.CounterVec
	gitlabDuration *prometheus.HistogramVec
}

// NewDefault calls [New] with [prometheus.DefaultRegisterer]
func NewDefault() (*Metrics, error) {
	return New(prometheus.DefaultRegisterer)
}

// New returns a metrics abstraction layer which registers
// all its internal metrics with the provided [prometheus.Registerer].
// All errors are the result of failed registrations.
func New(reg prometheus.Registerer) (*Metrics, error) {
	authFailures := prometheus.NewCounterVec(
		optsAuthFailures,
		[]string{labelRealm, labelCause},
	)
	authAttempts := prometheus.NewCounterVec(
		optsAuthAttempts,
		[]string{labelRealm},
	)
	gitlabDuration := prometheus.NewHistogramVec(
		optsGitlabDuration,
		[]string{labelService},
	)
	collectors := []prometheus.Collector{
		authFailures,
		authAttempts,
		gitlabDuration,
	}
	result := &Metrics{
		authFailures:   authFailures,
		authAttempts:   authAttempts,
		gitlabDuration: gitlabDuration,
	}

	for _, c := range collectors {
		if err := reg.Register(c); err != nil {
			return nil, err
		}
	}

	return result, nil
}

// AuthSuccess tracks a successful authentication.
func (m *Metrics) AuthSuccess(realm string) {
	m.authAttempts.With(prometheus.Labels{labelRealm: realm}).Inc()
}

// AuthMalformed tracks a failed authentication
// due to malformed user input.
func (m *Metrics) AuthMalformed(realm string) {
	m.authAttempts.With(prometheus.Labels{labelRealm: realm}).Inc()
	m.authFailures.With(prometheus.Labels{labelRealm: realm, labelCause: authCauseMalformed}).Inc()
}

// AuthNotFound tracks a failed authentication
// due to a lack of account information associated with the user input.
func (m *Metrics) AuthNotFound(realm string) {
	m.authAttempts.With(prometheus.Labels{labelRealm: realm}).Inc()
	m.authFailures.With(prometheus.Labels{labelRealm: realm, labelCause: authCauseNotFound}).Inc()
}

// AuthUnauthorized tracks a failed authentication
// due to the user not being authorized to
// access the provided realm.
func (m *Metrics) AuthUnauthorized(realm string) {
	m.authAttempts.With(prometheus.Labels{labelRealm: realm}).Inc()
	m.authFailures.With(prometheus.Labels{labelRealm: realm, labelCause: authCauseUnauthorized}).Inc()
}

// GitlabRequest reports on the elapsed time for the specific Gitlab service.
func (m *Metrics) GitlabRequest(service string, elapsed time.Duration) {
	m.gitlabDuration.With(prometheus.Labels{labelService: service}).Observe(elapsed.Seconds())
}
