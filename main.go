package main

import (
	"context"
	"fmt"
	"os/signal"
	"syscall"

	"log"

	"github.com/FACorreiaa/fitme-grpc/logger"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"os"

	"esgbook-software-engineer-technical-test-2024/internal/server"
	m "esgbook-software-engineer-technical-test-2024/middleware"

	"esgbook-software-engineer-technical-test-2024/middleware"
)

const timeout = 10

const file = "score_1.yaml"

func initializeLogger() error {
	return logger.Init(
		zap.DebugLevel,
		zap.String("service", "example"),
		zap.String("version", "v42.0.69"),
		zap.Strings("maintainers", []string{"@fc", "@FACorreiaa"}),
	)
}

func main() {
	if err := initializeLogger(); err != nil {
		panic("failed to initialize logging")
	}

	zapLogger := logger.Log
	grpcPort := os.Getenv("GRPC_SERVER_PORT")
	reg := prometheus.NewRegistry()
	println("Loaded prometheus registry")

	serverPort := os.Getenv("HTTP_SERVER_PORT")
	if serverPort == "" {
		serverPort = "8000"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("Got OS signal, shutting down gracefully...")
		cancel()
	}()

	// Initialize the multi-exporter (Tempo and Jaeger) and set the global tracer provider.
	if err := m.InitExporters(ctx); err != nil {
		zapLogger.Sugar().Error("Failed to initialize exporters", "err", err)
		log.Fatal(err)
	}

	errChan := make(chan error, 2)

	go func() {
		if err := server.RunHTTPServer(ctx, zapLogger, serverPort); err != nil {
			zapLogger.Sugar().Error("Failed to bootstrap server", "err", err)
			log.Fatal(err)
		}
	}()

	go func() {
		if err := middleware.ServePrometheus(ctx, ""); err != nil {
			zapLogger.Sugar().Error("Failed to serve prometheus", "err", err)
			log.Fatal(err)
		}
	}()

	go func() {
		if err := server.RunGRPCServer(ctx, zapLogger, grpcPort, reg); err != nil {
			zapLogger.Error("gRPC server error", zap.Error(err))
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		zapLogger.Info("Shutting down server")
	case err := <-errChan:
		zapLogger.Sugar().Error("Server exited with error", "err", err)
	}
}
