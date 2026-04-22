# Snowplow Metrics Reference

This document describes the OpenTelemetry metrics emitted by `snowplow`.

## Naming note

In code, metric names use dots (for example `snowplow.http.requests`).
In Prometheus, names are typically normalized with underscores (for example `snowplow_http_requests`), and counters may be exposed with `_total`.

## Metrics

| Metric | Type | Unit | Description | Emitted from | PromQL example |
|---|---|---|---|---|---|
| `snowplow.startup.success` | Counter | count | Service startup completed successfully. | `main.go` | `sum(increase(snowplow_startup_success_total[1h]))` |
| `snowplow.startup.failure` | Counter | count | Service terminated because the HTTP server failed after startup. | `main.go` | `sum(increase(snowplow_startup_failure_total[1h]))` |
| `snowplow.http.requests` | Counter | requests | Number of HTTP requests handled. Labels: `method`, `route`, `status_code`. | `internal/telemetry/metrics.go` | `sum by (route) (rate(snowplow_http_requests_total[5m]))` |
| `snowplow.http.duration_seconds` | Histogram | seconds | HTTP request latency. Labels: `method`, `route`, `status_code`. | `internal/telemetry/metrics.go` | `histogram_quantile(0.95, sum by (le, route) (rate(snowplow_http_duration_seconds_bucket[5m])))` |
| `snowplow.http.errors` | Counter | errors | HTTP requests completed with status >= 400. Labels: `method`, `route`, `status_code`. | `internal/telemetry/metrics.go` | `sum by (route) (rate(snowplow_http_errors_total[5m]))` |
| `snowplow.http.in_flight` | Gauge | requests | Current number of in-flight HTTP requests. | `internal/telemetry/metrics.go` | `max(snowplow_http_in_flight)` |
| `snowplow.call.requests` | Counter | requests | Number of `/call` requests. Labels: `method`, `api_group`, `resource`, `status_code`. | `internal/handlers/call.go` | `sum by (resource) (rate(snowplow_call_requests_total[5m]))` |
| `snowplow.call.duration_seconds` | Histogram | seconds | End-to-end `/call` latency. Labels: `method`, `api_group`, `resource`, `status_code`. | `internal/handlers/call.go` | `histogram_quantile(0.95, sum by (le, resource) (rate(snowplow_call_duration_seconds_bucket[5m])))` |
| `snowplow.call.validate.duration_seconds` | Histogram | seconds | Time spent validating and parsing `/call` request parameters. Labels: `stage`, `method`, `api_group`, `resource`. | `internal/handlers/call.go` | `histogram_quantile(0.95, sum by (le, resource) (rate(snowplow_call_validate_duration_seconds_bucket[5m])))` |
| `snowplow.call.build_uri.duration_seconds` | Histogram | seconds | Time spent building the upstream URI for `/call`. Labels: `stage`, `method`, `api_group`, `resource`. | `internal/handlers/call.go` | `histogram_quantile(0.95, sum by (le, resource) (rate(snowplow_call_build_uri_duration_seconds_bucket[5m])))` |
| `snowplow.call.user_config.duration_seconds` | Histogram | seconds | Time spent loading user configuration for `/call`. Labels: `stage`, `method`, `api_group`, `resource`. | `internal/handlers/call.go` | `histogram_quantile(0.95, sum by (le, resource) (rate(snowplow_call_user_config_duration_seconds_bucket[5m])))` |
| `snowplow.call.upstream.duration_seconds` | Histogram | seconds | Time spent in the upstream API invocation for `/call`. Labels: `stage`, `method`, `api_group`, `resource`. | `internal/handlers/call.go` | `histogram_quantile(0.95, sum by (le, resource) (rate(snowplow_call_upstream_duration_seconds_bucket[5m])))` |
| `snowplow.call.errors` | Counter | errors | Errors in `/call` flow. Labels: `stage`, `method`, `api_group`, `resource`, `status_code`. | `internal/handlers/call.go` | `sum by (stage, resource) (rate(snowplow_call_errors_total[5m]))` |
| `snowplow.list.requests` | Counter | requests | Number of `/list` requests. Labels: `category`, `status_code`. | `internal/handlers/list.go` | `sum by (category) (rate(snowplow_list_requests_total[5m]))` |
| `snowplow.list.duration_seconds` | Histogram | seconds | End-to-end `/list` latency. Labels: `category`, `status_code`. | `internal/handlers/list.go` | `histogram_quantile(0.95, sum by (le, category) (rate(snowplow_list_duration_seconds_bucket[5m])))` |
| `snowplow.list.discovery.duration_seconds` | Histogram | seconds | Time spent in category discovery for `/list`. Label: `category`. | `internal/handlers/list.go` | `histogram_quantile(0.95, sum by (le, category) (rate(snowplow_list_discovery_duration_seconds_bucket[5m])))` |
| `snowplow.list.resources_returned` | Counter | resources | Total resources returned by `/list`. Label: `category`. | `internal/handlers/list.go` | `sum(rate(snowplow_list_resources_returned_total[5m]))` |
| `snowplow.list.errors` | Counter | errors | Errors in `/list` flow. Labels: `category`, `stage`, `status_code`. | `internal/handlers/list.go` | `sum by (stage, category) (rate(snowplow_list_errors_total[5m]))` |
| `snowplow.jq.requests` | Counter | requests | Number of `/jq` requests. Label: `status_code`. | `internal/handlers/jq.go` | `sum(rate(snowplow_jq_requests_total[5m]))` |
| `snowplow.jq.duration_seconds` | Histogram | seconds | End-to-end `/jq` latency. Label: `status_code`. | `internal/handlers/jq.go` | `histogram_quantile(0.95, sum by (le) (rate(snowplow_jq_duration_seconds_bucket[5m])))` |
| `snowplow.jq.decode.duration_seconds` | Histogram | seconds | Time spent decoding the `/jq` request body. | `internal/handlers/jq.go` | `histogram_quantile(0.95, sum by (le) (rate(snowplow_jq_decode_duration_seconds_bucket[5m])))` |
| `snowplow.jq.eval.duration_seconds` | Histogram | seconds | Time spent evaluating the JQ expression. | `internal/handlers/jq.go` | `histogram_quantile(0.95, sum by (le) (rate(snowplow_jq_eval_duration_seconds_bucket[5m])))` |
| `snowplow.jq.errors` | Counter | errors | Errors in `/jq` flow. Labels: `stage`, `status_code`. | `internal/handlers/jq.go` | `sum by (stage) (rate(snowplow_jq_errors_total[5m]))` |

## Cardinality guidance

- Avoid high-cardinality labels like user names, object names, resource IDs, or query values.
- The current `route` label is normalized for fixed endpoints and Swagger assets (`/swagger/*`).
- Current labels are bounded: `method`, `route`, `status_code`.
