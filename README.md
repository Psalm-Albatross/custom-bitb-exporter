# Custom Bitbucket Exporter

This exporter collects critical Bitbucket Server/Data Center metrics and exposes them for Prometheus scraping. Visualize and alert on these metrics using Grafana.

## Features
- Repository count (more metrics can be added)
- Easy integration with Prometheus and Grafana

## Setup
1. Set environment variables:
   - `BITBUCKET_URL` (e.g., https://bitbucket.example.com)
   - `BITBUCKET_USERNAME`
   - `BITBUCKET_PASSWORD`
2. Build and run:
   ```sh
   go build -o bitb-exporter
   ./bitb-exporter
   ```
3. Prometheus scrape config:
   ```yaml
   - job_name: 'bitbucket'
     static_configs:
       - targets: ['localhost:8080']
   ```

## Extending
Add more collectors in `collector.go` and API calls in `bitbucket_client.go` for additional metrics (PRs, users, system health, etc).
