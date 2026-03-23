package observability

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitTracer initializes an OpenTelemetry tracer provider pipeline.
// In a real environment, you'd configure an OTLP exporter here to send to Jaeger/Zipkin/Honeycomb.
func InitTracer(serviceName string) (*sdktrace.TracerProvider, error) {
	res, err := resource.New(context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
		),
	)
	if err != nil {
		return nil, err
	}

	// We use a Noop exporter for the demo since we don't assume a local Jaeger instance is running
	// Normally: exporter, err := otlptracegrpc.New(ctx, ...)
	
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	slog.Info("Initialized OpenTelemetry distributed tracing", "service", serviceName)
	return tp, nil
}
