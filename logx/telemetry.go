package logx

import (
	"context"
	"errors"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// InitTelemetry configures W3C propagation and OTLP HTTP export to Jaeger.
func InitTelemetry(ctx context.Context, service string) (func(context.Context) error, error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return func(context.Context) error { return nil }, err
	}
	res, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(semconv.ServiceName(service)),
	)
	if err != nil {
		return func(context.Context) error { return nil }, err
	}
	provider := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(provider)

	return func(shutdownCtx context.Context) error {
		if err := provider.Shutdown(shutdownCtx); err != nil && !errors.Is(err, context.Canceled) {
			return err
		}
		return nil
	}, nil
}

// Init starts telemetry without making startup depend on the collector.
func Init(ctx context.Context, service string) func(context.Context) error {
	shutdown, err := InitTelemetry(ctx, service)
	if err != nil {
		slog.Warn("otel setup failed", "err", err)
		return func(context.Context) error { return nil }
	}
	return shutdown
}
