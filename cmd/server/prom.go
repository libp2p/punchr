package main

import "github.com/prometheus/client_golang/prometheus"

var allocationQueryDurationHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: "db_allocation_query_duration_seconds",
	Help: "Histogram of database query times for client allocations",
}, []string{"success"})

func init() {
	prometheus.MustRegister(allocationQueryDurationHistogram)
}
