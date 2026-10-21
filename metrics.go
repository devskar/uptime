package main

import (
	"errors"
	"log"
	"net"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	checkUp = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "uptime_check_up",
		Help: "1 if the last check of the target succeeded, 0 otherwise.",
	}, []string{"target"})

	checkDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "uptime_check_duration_seconds",
		Help:    "Duration of checks that received a response.",
		Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"target"})

	checkRuns = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "uptime_check_runs_total",
		Help: "Total number of checks by outcome.",
	}, []string{"target", "outcome"})

	checkStatusCode = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "uptime_check_last_status_code",
		Help: "HTTP status code of the last response.",
	}, []string{"target"})
)

func recordResult(r HealthcheckResult) {
	log.Println("Recording result")
	checkRuns.WithLabelValues(r.URL, outcome(r)).Inc()
	checkUp.WithLabelValues(r.URL).Set(checkupValue(r))

	if r.StatusCode != 0 {
		checkDuration.WithLabelValues(r.URL).Observe(r.Duration.Seconds())
		checkStatusCode.WithLabelValues(r.URL).Set(float64(r.StatusCode))
	}

}

func checkupValue(r HealthcheckResult) float64 {
	if r.Up {
		return 1
	}
	return 0
}

func outcome(r HealthcheckResult) string {
	if r.Up {
		return "success"
	}

	var dnsErr *net.DNSError
	var netErr net.Error

	switch {
	case r.Err == nil:
		return "status"
	case errors.As(r.Err, &dnsErr):
		return "dns"
	case errors.As(r.Err, &netErr) && netErr.Timeout():
		return "timeout"
	case r.StatusCode != 0:
		return "body"
	default:
		return "connection"
	}
}
