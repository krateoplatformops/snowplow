# Telemetry Assets

This folder contains ready-to-use telemetry assets for `snowplow`:

- `dashboards/snowplow-overview.dashboard.json`: Grafana dashboard with example panels
- `collector/otel-collector-config.yaml`: minimal OpenTelemetry Collector config (OTLP HTTP -> Prometheus endpoint)
- `metrics-reference.md`: full metric catalog (type, meaning, examples)

## Prerequisites

1. `snowplow` running with OpenTelemetry enabled:

```yaml
env:
  OTEL_ENABLED: "true"
  OTEL_EXPORT_INTERVAL: "30s"
  OTEL_EXPORTER_OTLP_ENDPOINT: "http://otel-collector.monitoring.svc.cluster.local:4318"
```

2. OpenTelemetry Collector reachable from `snowplow`
3. Prometheus scraping the Collector Prometheus exporter (default in this example: `:9464`)
4. Grafana connected to Prometheus as a data source

## Import The Dashboard

1. Open Grafana
2. Go to `Dashboards` -> `New` -> `Import`
3. Upload `dashboards/snowplow-overview.dashboard.json`
4. Select your Prometheus data source
5. Save

## Example Panels Included

- Startup success/failure counters (last hour)
- HTTP request rate by route
- HTTP error rate by route
- HTTP latency (p50/p95)
- In-flight requests

## Metric Naming Notes

Depending on your OTel -> Prometheus conversion rules:

- counters may appear as `<metric>_total`
- histograms usually appear as `<metric>_bucket`, `<metric>_sum`, `<metric>_count`

The provided dashboard uses `or` fallbacks for many counters (for example `metric_total or metric`) to reduce friction.
If your environment still differs, edit panel queries accordingly.

## Collector Example

You can start from `collector/otel-collector-config.yaml` and adapt it to your deployment.

Current pipeline in the example:

- Receiver: OTLP HTTP on `4318`
- Processor: `batch`
- Exporter: Prometheus endpoint on `9464`
