package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

type metricsData struct {
	registry *prometheus.Registry
	// goregistry      gometrics.Registry
	sendTotal       *prometheus.CounterVec
	sendSizes       *prometheus.SummaryVec
	receiveTotal    *prometheus.CounterVec
	receiveSizes    *prometheus.SummaryVec
	receiveDuration *prometheus.HistogramVec
}

var metrics metricsData

var durationBucket = []float64{0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10}

func (m *metricsData) Init() http.Handler {
	m.sendTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "melp_send_total",
			Help: "Tracks the number of sent messages.",
		}, []string{"topic", "partition"},
	)
	m.sendSizes = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "melp_send_size_bytes",
			Help: "Tracks the number of bytes sent",
		}, []string{"topic", "partition"},
	)
	m.receiveTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "melp_receive_total",
			Help: "Tracks the number of received messages.",
		}, []string{"topic", "partition", "status"},
	)
	m.receiveSizes = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "melp_receive_size_bytes",
			Help: "Tracks the number of bytes received",
		}, []string{"topic", "partition", "status"},
	)
	m.receiveDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "melp_receive_duration_seconds",
			Help:    "Tracks the latencies of received messages",
			Buckets: durationBucket,
		}, []string{"topic", "partition", "status"},
	)

	if config.Metrics.Go {
		m.registry.MustRegister(collectors.NewGoCollector())
	}
	if config.Metrics.Process {
		m.registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	}

	m.registry = prometheus.NewRegistry()
	m.registry.MustRegister(
		metrics.sendTotal,
		metrics.sendSizes,
		metrics.receiveTotal,
		metrics.receiveSizes,
		metrics.receiveDuration,
	)

	// m.goregistry = gometrics.DefaultRegistry

	// kafkaMetricsCollector := prommetrics.NewPrometheusProvider(m.goregistry, "melp", "kafka", m.registry, 1*time.Second)
	// go kafkaMetricsCollector.UpdatePrometheusMetrics()

	//return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{Registry: m.registry})
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}

func (m *metricsData) Send(topic string, partition int32, size int) {
	if m.sendTotal == nil {
		return
	}
	p := strconv.Itoa(int(partition))
	m.sendTotal.WithLabelValues(topic, p).Inc()
	m.sendSizes.WithLabelValues(topic, p).Observe(float64(size))
}

func (m *metricsData) Receive(topic string, partition int32, size int, dur time.Duration, status string) {
	if m.receiveTotal == nil {
		return
	}
	p := strconv.Itoa(int(partition))
	m.receiveTotal.WithLabelValues(topic, p, status).Inc()
	m.receiveSizes.WithLabelValues(topic, p, status).Observe(float64(size))
	m.receiveDuration.WithLabelValues(topic, p, status).Observe(dur.Seconds())
}
