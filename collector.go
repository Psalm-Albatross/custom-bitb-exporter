package main

import (
	"github.com/prometheus/client_golang/prometheus"
)

type BitbucketCollector struct {
	client    *BitbucketClient
	repoCount *prometheus.Desc
}

func NewBitbucketCollector(client *BitbucketClient) *BitbucketCollector {
	return &BitbucketCollector{
		client:    client,
		repoCount: prometheus.NewDesc("bitbucket_repository_count", "Total number of repositories", nil, nil),
	}
}

func (c *BitbucketCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.repoCount
}

func (c *BitbucketCollector) Collect(ch chan<- prometheus.Metric) {
	repoCount, err := c.client.GetRepositoryCount()
	if err == nil {
		ch <- prometheus.MustNewConstMetric(c.repoCount, prometheus.GaugeValue, float64(repoCount))
	}
}
