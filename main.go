package main

import (
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Create Bitbucket client
	client := NewBitbucketClient(cfg)

	// Register Prometheus collector
	collector := NewBitbucketCollector(client)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Exporter running on :8080/metrics")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
