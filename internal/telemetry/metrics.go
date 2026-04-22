package telemetry

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
)

type Config struct {
	Enabled        bool
	ServiceName    string
	ExportInterval time.Duration
}

type Metrics struct {
	startupSuccess metric.Int64Counter
	startupFailure metric.Int64Counter

	httpRequests  metric.Int64Counter
	httpDuration  metric.Float64Histogram
	httpErrors    metric.Int64Counter
	httpInFlight  metric.Int64ObservableGauge
	inFlightCount atomic.Int64

	callRequests           metric.Int64Counter
	callDuration           metric.Float64Histogram
	callValidateDuration   metric.Float64Histogram
	callBuildURIDuration   metric.Float64Histogram
	callUserConfigDuration metric.Float64Histogram
	callUpstreamDuration   metric.Float64Histogram
	callErrors             metric.Int64Counter

	listRequests          metric.Int64Counter
	listDuration          metric.Float64Histogram
	listDiscoveryDuration metric.Float64Histogram
	listResourcesReturned metric.Int64Counter
	listErrors            metric.Int64Counter

	jqRequests       metric.Int64Counter
	jqDuration       metric.Float64Histogram
	jqDecodeDuration metric.Float64Histogram
	jqEvalDuration   metric.Float64Histogram
	jqErrors         metric.Int64Counter
}

func Setup(ctx context.Context, log *slog.Logger, cfg Config) (*Metrics, func(context.Context) error, error) {
	if !cfg.Enabled {
		return nil, func(context.Context) error { return nil }, nil
	}

	exporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return nil, nil, err
	}

	res, err := resource.Merge(resource.Default(),
		resource.NewSchemaless(attribute.String("service.name", cfg.ServiceName)))
	if err != nil {
		return nil, nil, err
	}

	readerOptions := []sdkmetric.PeriodicReaderOption{}
	if cfg.ExportInterval > 0 {
		readerOptions = append(readerOptions, sdkmetric.WithInterval(cfg.ExportInterval))
	}

	reader := sdkmetric.NewPeriodicReader(exporter, readerOptions...)
	provider := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(provider)

	meter := provider.Meter("github.com/krateoplatformops/snowplow")
	metrics, err := newMetrics(meter)
	if err != nil {
		_ = provider.Shutdown(ctx)
		return nil, nil, err
	}

	log.Info("OpenTelemetry metrics initialized")
	return metrics, provider.Shutdown, nil
}

func newMetrics(meter metric.Meter) (*Metrics, error) {
	var err error
	m := &Metrics{}

	if m.startupSuccess, err = meter.Int64Counter("snowplow.startup.success"); err != nil {
		return nil, err
	}
	if m.startupFailure, err = meter.Int64Counter("snowplow.startup.failure"); err != nil {
		return nil, err
	}
	if m.httpRequests, err = meter.Int64Counter("snowplow.http.requests"); err != nil {
		return nil, err
	}
	if m.httpDuration, err = meter.Float64Histogram("snowplow.http.duration_seconds"); err != nil {
		return nil, err
	}
	if m.httpErrors, err = meter.Int64Counter("snowplow.http.errors"); err != nil {
		return nil, err
	}
	if m.httpInFlight, err = meter.Int64ObservableGauge("snowplow.http.in_flight"); err != nil {
		return nil, err
	}
	if m.callRequests, err = meter.Int64Counter("snowplow.call.requests"); err != nil {
		return nil, err
	}
	if m.callDuration, err = meter.Float64Histogram("snowplow.call.duration_seconds"); err != nil {
		return nil, err
	}
	if m.callValidateDuration, err = meter.Float64Histogram("snowplow.call.validate.duration_seconds"); err != nil {
		return nil, err
	}
	if m.callBuildURIDuration, err = meter.Float64Histogram("snowplow.call.build_uri.duration_seconds"); err != nil {
		return nil, err
	}
	if m.callUserConfigDuration, err = meter.Float64Histogram("snowplow.call.user_config.duration_seconds"); err != nil {
		return nil, err
	}
	if m.callUpstreamDuration, err = meter.Float64Histogram("snowplow.call.upstream.duration_seconds"); err != nil {
		return nil, err
	}
	if m.callErrors, err = meter.Int64Counter("snowplow.call.errors"); err != nil {
		return nil, err
	}
	if m.listRequests, err = meter.Int64Counter("snowplow.list.requests"); err != nil {
		return nil, err
	}
	if m.listDuration, err = meter.Float64Histogram("snowplow.list.duration_seconds"); err != nil {
		return nil, err
	}
	if m.listDiscoveryDuration, err = meter.Float64Histogram("snowplow.list.discovery.duration_seconds"); err != nil {
		return nil, err
	}
	if m.listResourcesReturned, err = meter.Int64Counter("snowplow.list.resources_returned"); err != nil {
		return nil, err
	}
	if m.listErrors, err = meter.Int64Counter("snowplow.list.errors"); err != nil {
		return nil, err
	}
	if m.jqRequests, err = meter.Int64Counter("snowplow.jq.requests"); err != nil {
		return nil, err
	}
	if m.jqDuration, err = meter.Float64Histogram("snowplow.jq.duration_seconds"); err != nil {
		return nil, err
	}
	if m.jqDecodeDuration, err = meter.Float64Histogram("snowplow.jq.decode.duration_seconds"); err != nil {
		return nil, err
	}
	if m.jqEvalDuration, err = meter.Float64Histogram("snowplow.jq.eval.duration_seconds"); err != nil {
		return nil, err
	}
	if m.jqErrors, err = meter.Int64Counter("snowplow.jq.errors"); err != nil {
		return nil, err
	}

	_, err = meter.RegisterCallback(func(_ context.Context, observer metric.Observer) error {
		observer.ObserveInt64(m.httpInFlight, m.inFlightCount.Load())
		return nil
	}, m.httpInFlight)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Metrics) IncStartupSuccess(ctx context.Context) {
	if m == nil {
		return
	}
	m.startupSuccess.Add(ctx, 1)
}

func (m *Metrics) IncStartupFailure(ctx context.Context) {
	if m == nil {
		return
	}
	m.startupFailure.Add(ctx, 1)
}

func (m *Metrics) WrapHTTP(next http.Handler) http.Handler {
	if m == nil {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		m.inFlightCount.Add(1)
		defer m.inFlightCount.Add(-1)

		rec := &statusRecorder{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rec, r)

		route := normalizeRoute(r.URL.Path)
		m.RecordHTTPRequest(r.Context(), r.Method, route, rec.statusCode, time.Since(start))
		if rec.statusCode >= http.StatusBadRequest {
			m.IncHTTPError(r.Context(), r.Method, route, rec.statusCode)
		}
	})
}

func (m *Metrics) RecordHTTPRequest(ctx context.Context, method, route string, status int, d time.Duration) {
	if m == nil {
		return
	}

	attrs := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("route", route),
		attribute.Int("status_code", status),
	)
	m.httpRequests.Add(ctx, 1, attrs)
	m.httpDuration.Record(ctx, d.Seconds(), attrs)
}

func (m *Metrics) IncHTTPError(ctx context.Context, method, route string, status int) {
	if m == nil {
		return
	}

	m.httpErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("route", route),
		attribute.Int("status_code", status),
	))
}

func (m *Metrics) RecordCallRequest(ctx context.Context, method, apiGroup, resource string, status int, d time.Duration) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("method", method),
		attribute.String("api_group", apiGroup),
		attribute.String("resource", resource),
		attribute.Int("status_code", status),
	)
	m.callRequests.Add(ctx, 1, attrs)
	m.callDuration.Record(ctx, d.Seconds(), attrs)
}

func (m *Metrics) RecordCallStageDuration(ctx context.Context, stage, method, apiGroup, resource string, d time.Duration) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("stage", stage),
		attribute.String("method", method),
		attribute.String("api_group", apiGroup),
		attribute.String("resource", resource),
	)
	switch stage {
	case "validate":
		m.callValidateDuration.Record(ctx, d.Seconds(), attrs)
	case "build_uri":
		m.callBuildURIDuration.Record(ctx, d.Seconds(), attrs)
	case "user_config":
		m.callUserConfigDuration.Record(ctx, d.Seconds(), attrs)
	case "upstream":
		m.callUpstreamDuration.Record(ctx, d.Seconds(), attrs)
	}
}

func (m *Metrics) IncCallError(ctx context.Context, stage, method, apiGroup, resource string, status int) {
	if m == nil {
		return
	}
	m.callErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("stage", stage),
		attribute.String("method", method),
		attribute.String("api_group", apiGroup),
		attribute.String("resource", resource),
		attribute.Int("status_code", status),
	))
}

func (m *Metrics) RecordListRequest(ctx context.Context, category string, status int, d time.Duration) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(
		attribute.String("category", category),
		attribute.Int("status_code", status),
	)
	m.listRequests.Add(ctx, 1, attrs)
	m.listDuration.Record(ctx, d.Seconds(), attrs)
}

func (m *Metrics) RecordListDiscoveryDuration(ctx context.Context, category string, d time.Duration) {
	if m == nil {
		return
	}
	m.listDiscoveryDuration.Record(ctx, d.Seconds(),
		metric.WithAttributes(attribute.String("category", category)))
}

func (m *Metrics) AddListResourcesReturned(ctx context.Context, category string, n int64) {
	if m == nil || n <= 0 {
		return
	}
	m.listResourcesReturned.Add(ctx, n,
		metric.WithAttributes(attribute.String("category", category)))
}

func (m *Metrics) IncListError(ctx context.Context, category, stage string, status int) {
	if m == nil {
		return
	}
	m.listErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("category", category),
		attribute.String("stage", stage),
		attribute.Int("status_code", status),
	))
}

func (m *Metrics) RecordJQRequest(ctx context.Context, status int, d time.Duration) {
	if m == nil {
		return
	}
	attrs := metric.WithAttributes(attribute.Int("status_code", status))
	m.jqRequests.Add(ctx, 1, attrs)
	m.jqDuration.Record(ctx, d.Seconds(), attrs)
}

func (m *Metrics) RecordJQDecodeDuration(ctx context.Context, d time.Duration) {
	if m == nil {
		return
	}
	m.jqDecodeDuration.Record(ctx, d.Seconds())
}

func (m *Metrics) RecordJQEvalDuration(ctx context.Context, d time.Duration) {
	if m == nil {
		return
	}
	m.jqEvalDuration.Record(ctx, d.Seconds())
}

func (m *Metrics) IncJQError(ctx context.Context, stage string, status int) {
	if m == nil {
		return
	}
	m.jqErrors.Add(ctx, 1, metric.WithAttributes(
		attribute.String("stage", stage),
		attribute.Int("status_code", status),
	))
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	return r.ResponseWriter.Write(data)
}

func normalizeRoute(path string) string {
	switch {
	case strings.HasPrefix(path, "/swagger/"):
		return "/swagger/*"
	case path == "":
		return "/"
	default:
		return path
	}
}
