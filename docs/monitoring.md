# Monitoring

Monitoring the application allows humans and automated system to make informed
decisions related to scaling, failover and protection mechanisms.

# Health checks

Using the `health.port` configuration setting, instructs the application to
listen on the provided port for HTTP requests. Two endpoints are available:
`/-/health` and `/-/ready`. They are intended for scheduling system such
as [Kubernetes][].

[Kubernetes]: https://kubernetes.io/docs/concepts/configuration/liveness-readiness-startup-probes/

# Metrics

Application performance information is available on a dedicated port if
the `metrics.port` configuration setting is defined. Data is accessible
under `/metrics` as [Prometheus][] exposition format.

[Prometheus]: https://prometheus.io/docs/instrumenting/exposition_formats/

The following metrics are available:

| Metric                                              | Type         | Description                                                         |
|-----------------------------------------------------|--------------|---------------------------------------------------------------------|
| gitlab_authn_build_info                             | gauge        | Application information                                             |
| gitlab_authn_authentication_attempts_total          | counter      | Number of authentication attempts.                                  |
| gitlab_authn_authentication_failures_total          | counter      | Number of authentication failures.                                  |
| gitlab_authn_userinfo_cache_evictions_total         | counter      | Number of items removed from the cache.                             |
| gitlab_authn_userinfo_cache_hits_total              | counter      | Number of successful retrievals.                                    |
| gitlab_authn_userinfo_cache_insertions_total        | counter      | Number of inserted items.                                           |
| gitlab_authn_userinfo_cache_misses_total            | counter      | Number of items which where not found.                              |
| gitlab_authn_gitlab_request_duration_seconds        | histogram    | Elapsed time in seconds for HTTP request against Gitlab.            |

# Profiling

The application exposes various profiling endpoints when enabled with `profile.port`:

| Endpoint                        | Description                                                                                                 |
|---------------------------------|-------------------------------------------------------------------------------------------------------------|
| `/debug/pprof/`                 | [Profile overview](https://pkg.go.dev/net/http/pprof#Index)                                                 |
| `/debug/pprof/goroutine`        | [Stack traces of all current goroutines](https://pkg.go.dev/runtime/pprof#Profile)                          |
| `/debug/pprof/heap`             | [Sampling of memory allocations of live objects](https://pkg.go.dev/runtime/pprof#Profile)                  |
| `/debug/pprof/allocs`           | [Sampling of all past memory allocations](https://pkg.go.dev/runtime/pprof#Profile)                         |
| `/debug/pprof/threadcreate`     | [Stack traces that led to the creation of new OS threads](https://pkg.go.dev/runtime/pprof#Profile)         |
| `/debug/pprof/block`            | [Stack traces that led to blocking on synchronization primitives](https://pkg.go.dev/runtime/pprof#Profile) |
| `/debug/pprof/mutex`            | [Stack traces of holders of contended mutexes](https://pkg.go.dev/runtime/pprof#Profile)                    |
| `/debug/pprof/cmdline`          | [Application commandline](https://pkg.go.dev/net/http/pprof#Cmdline)                                        |
| `/debug/pprof/profile`          | [CPU profiler](https://pkg.go.dev/net/http/pprof#Profile)                                                   |
| `/debug/pprof/symbol`           | [Program counter mapping](https://pkg.go.dev/net/http/pprof#Symbol)                                         |
| `/debug/pprof/trace`            | [Execution tracer](https://pkg.go.dev/net/http/pprof#Trace)                                                 |
| `/debug/pprof/delta_heap`       | [Heap analysis](https://pkg.go.dev/github.com/grafana/pyroscope-go/godeltaprof/http/pprof#Heap)             |
| `/debug/pprof/delta_block`      | [Profile overview](https://pkg.go.dev/github.com/grafana/pyroscope-go/godeltaprof/http/pprof#Block)         |
| `/debug/pprof/delta_mutex`      | [Mutext statistics](https://pkg.go.dev/github.com/grafana/pyroscope-go/godeltaprof/http/pprof#Mutex)        |

The *delta_* endpoints report incremental values instead of the full report which grows over time.
Data can be scraped using [Grafana Alloy][] for further processing by a Tracing/Profiling stack.

[Grafana Alloy]: https://grafana.com/docs/alloy/latest/reference/components/pyroscope/pyroscope.scrape/#pyroscopescrape

