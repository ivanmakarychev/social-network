package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	_ "github.com/prometheus/client_golang/prometheus/promauto"
)

type (
	Metrics interface {
		CountRequest(m *RequestMetrics)
	}

	RequestMetrics struct {
		URL      string
		HTTPCode int
		Duration float64
	}

	Impl struct {
		handledRequests *prometheus.HistogramVec
	}
)

func (i *Impl) CountRequest(m *RequestMetrics) {
	i.handledRequests.WithLabelValues(m.URL, strconv.Itoa(m.HTTPCode)).Observe(m.Duration)
}

func New() *Impl {
	return &Impl{
		handledRequests: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dialogues_handled_requests",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"handler", "http_code"},
		),
	}
}
