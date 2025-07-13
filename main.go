package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	port := flag.String("port", "8080", "Port to listen on")
	logLevel := flag.String("log.level", "info", "Log level: debug, info, warn, error")
	cloud := flag.Bool("cloud", false, "Set to true for Bitbucket Cloud, false for Data Center/Server")
	flag.Parse()

	log.Printf("Starting Bitbucket exporter on :%s/metrics (log level: %s, cloud: %v)", *port, *logLevel, *cloud)

	// Load config
	cfg, err := LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Create Bitbucket client
	client := NewBitbucketClient(cfg, *cloud)

	// Register Prometheus collector
	collector := NewBitbucketCollector(client, *logLevel)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+*port, nil))
}
