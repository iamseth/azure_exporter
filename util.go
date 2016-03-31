package main

import (
	"log"
	"os/user"

	"github.com/prometheus/client_golang/prometheus"
)

// newGuageVec is a small helper function to create a new *prometheus.GaugeVec.
func newGuageVec(metricsName, docString string, labels []string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      metricsName,
			Help:      docString,
		},
		labels,
	)
}

// newCounterVec is a small helper function to create a new *prometheus.CounterVec.
func newCounterVec(metricsName, docString string, labels []string) *prometheus.CounterVec {
	return prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      metricsName,
			Help:      docString,
		},
		labels,
	)
}

// Return the full path of the current user's home directory.
func getHomeDirectory() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("Unable to determine user's home directory: %s", err)
	}
	return usr.HomeDir
}
