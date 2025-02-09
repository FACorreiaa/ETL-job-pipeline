package middleware

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	oteltrace "go.opentelemetry.io/otel/sdk/trace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// NewConsoleExporter Console Exporter, only for testing locally
func NewConsoleExporter() (oteltrace.SpanExporter, error) {
	return stdouttrace.New()
}

type multiExporter struct {
	exporters []sdktrace.SpanExporter
}

func (m *multiExporter) Shutdown(ctx context.Context) error {
	var lastErr error
	for _, exp := range m.exporters {
		if err := exp.Shutdown(ctx); err != nil {
			// You could choose to combine errors or log them.
			lastErr = err
		}
	}
	return lastErr
}

func NewMultiExporter(exporters ...sdktrace.SpanExporter) sdktrace.SpanExporter {
	return &multiExporter{exporters: exporters}
}

func (m *multiExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	var lastErr error
	for _, exp := range m.exporters {
		if err := exp.ExportSpans(ctx, spans); err != nil {
			// You could choose to combine errors or log them.
			lastErr = err
		}
	}
	return lastErr
}

// NewOTLPExporter OTLP Exporter
func NewOTLPExporter(ctx context.Context, endpoint string) (oteltrace.SpanExporter, error) {
	if endpoint == "" {
		endpoint = "tempo:4318" // default if none provided
	}

	insecureOpt := otlptracehttp.WithInsecure()
	endpointOpt := otlptracehttp.WithEndpoint(endpoint)
	pathOpt := otlptracehttp.WithURLPath("/v1/traces")

	return otlptracehttp.New(ctx, insecureOpt, endpointOpt, pathOpt)
}

//func NewTraceProvider(exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
//	r, err := resource.Merge(
//		resource.Default(),
//		resource.NewWithAttributes(
//			semconv.SchemaURL,
//			semconv.ServiceName("score-app"),
//		),
//	)
//
//	if err != nil {
//		panic(err)
//	}
//
//	return sdktrace.NewTracerProvider(
//		sdktrace.WithBatcher(exp),
//		sdktrace.WithResource(r),
//	)
//}

func InitExporters(ctx context.Context) error {
	tempoExporter, err := NewOTLPExporter(ctx, "tempo:4318")
	if err != nil {
		return fmt.Errorf("failed to create Tempo exporter: %w", err)
	}

	jaegerExporter, err := NewOTLPExporter(ctx, "jaeger:4318")
	if err != nil {
		return fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	multiExporter := NewMultiExporter(tempoExporter, jaegerExporter)

	// Create a tracer provider using the multi-exporter.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(multiExporter),
		// Add any additional options or resources as needed.
	)
	otel.SetTracerProvider(tp)

	// Optionally, you can store tp so that you can call Shutdown later.
	return nil
}
