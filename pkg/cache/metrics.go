package cache

import (
	ttlcache "github.com/jellydator/ttlcache/v3"

	"github.com/prometheus/client_golang/prometheus"
)

type metricsCollector struct {
	source func() ttlcache.Metrics

	insertions *prometheus.Desc
	hits       *prometheus.Desc
	misses     *prometheus.Desc
	evictions  *prometheus.Desc
}

func NewMetricsCollector(source func() ttlcache.Metrics, namespace string) prometheus.Collector {
	subsystem := "userinfo_cache"

	insertions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "insertions_total"),
		"Number of inserted items.",
		nil, nil)
	hits := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "hits_total"),
		"Number of successful retrievals.",
		nil, nil)
	misses := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "misses_total"),
		"Number of items which where not found.",
		nil, nil)
	evictions := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, "evictions_total"),
		"Number of items removed from the cache.",
		nil, nil)

	result := &metricsCollector{
		source:     source,
		insertions: insertions,
		hits:       hits,
		misses:     misses,
		evictions:  evictions,
	}

	return result
}

func (c *metricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.insertions
	ch <- c.hits
	ch <- c.misses
	ch <- c.evictions
}

func (c *metricsCollector) Collect(ch chan<- prometheus.Metric) {
	m := c.source()

	ch <- prometheus.MustNewConstMetric(c.insertions, prometheus.CounterValue, float64(m.Insertions))
	ch <- prometheus.MustNewConstMetric(c.hits, prometheus.CounterValue, float64(m.Hits))
	ch <- prometheus.MustNewConstMetric(c.misses, prometheus.CounterValue, float64(m.Misses))
	ch <- prometheus.MustNewConstMetric(c.evictions, prometheus.CounterValue, float64(m.Evictions))
}
