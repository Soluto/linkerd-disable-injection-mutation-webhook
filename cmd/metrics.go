package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	mutationErrors = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mutation_errors",
		Help: "Mutation errors",
	})
)

// InitMetrics register the metrics
func InitMetrics() {
	prometheus.MustRegister(mutationErrors)
}
