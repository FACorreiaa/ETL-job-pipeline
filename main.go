package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"esgbook-software-engineer-technical-test-2024/internal/scoring"
	"esgbook-software-engineer-technical-test-2024/middleware"
)

const timeout = 10

const file = "score_1.yaml"

func initExporters(ctx context.Context) error {
	tempoExporter, err := middleware.NewOTLPExporter(ctx, "tempo:4318")
	if err != nil {
		return fmt.Errorf("failed to create Tempo exporter: %w", err)
	}

	jaegerExporter, err := middleware.NewOTLPExporter(ctx, "jaeger:4318")
	if err != nil {
		return fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	multiExporter := middleware.NewMultiExporter(tempoExporter, jaegerExporter)

	// Create a tracer provider using the multi-exporter.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(multiExporter),
		// Add any additional options or resources as needed.
	)
	otel.SetTracerProvider(tp)

	// Optionally, you can store tp so that you can call Shutdown later.
	return nil
}

func BoostrapServer(logger *slog.Logger, ctx context.Context) error {
	server := http.NewServeMux()
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8000"
	}

	h := scoring.Handler{
		Ctx:            ctx,
		Logger:         logger,
		ConfigFileName: file,
	}

	// No need to create a new exporter or trace provider here.
	// The global tracer provider has already been set in main by initExporters.

	server.HandleFunc("/run-scores", h.CalculateScoreHandler)
	server.HandleFunc("/health", scoring.HealthCheckHandler)

	wrapped := middleware.LoggingMiddleware(logger)(server)
	logger.Info("Starting service on :" + serverPort)
	if err := http.ListenAndServe(":"+serverPort, wrapped); err != nil {
		logger.Error("Failed to start server", "err", err)
		return err
	}
	return nil
}

func main() {
	logger := middleware.InitLogger()
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Minute)
	defer cancel()

	// Initialize the multi-exporter (Tempo and Jaeger) and set the global tracer provider.
	if err := initExporters(ctx); err != nil {
		logger.Error("Failed to initialize exporters", "err", err)
		log.Fatal(err)
	}

	// Start your server and Prometheus concurrently.
	errChan := make(chan error, 2)

	go func() {
		if err := BoostrapServer(logger, ctx); err != nil {
			logger.Error("Failed to bootstrap server", "err", err)
			log.Fatal(err)
		}
	}()

	go func() {
		if err := middleware.ServePrometheus(ctx, ""); err != nil {
			logger.Error("Failed to serve prometheus", "err", err)
			log.Fatal(err)
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutting down server")
	case err := <-errChan:
		logger.Error("Server exited with error", "err", err)
	}
}
