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

