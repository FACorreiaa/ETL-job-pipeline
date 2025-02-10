package main

import (
	"context"
	"log"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"esgbook-software-engineer-technical-test-2024/middleware"

	"os"

	"esgbook-software-engineer-technical-test-2024/internal/server"
	m "esgbook-software-engineer-technical-test-2024/middleware"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	zapLogger, err := middleware.InitializeLogger()
	if err != nil {
		panic("failed to initialize logging")
	}

	grpcPort := os.Getenv("GRPC_SERVER_PORT")
	if grpcPort == "" {
		grpcPort = "8001"
	}
	reg := prometheus.NewRegistry()
	println("Loaded prometheus registry")

	serverPort := os.Getenv("HTTP_SERVER_PORT")
	if serverPort == "" {
		serverPort = "8000"
	}

	//c := make(chan os.Signal, 1)
	//signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	//go func() {
	//	<-c
	//	fmt.Println("Got OS signal, shutting down gracefully...")
	//	cancel()
	//}()

	// Initialize the multi-exporter (Tempo and Jaeger) and set the global tracer provider.
	if err := m.InitExporters(ctx); err != nil {
		zapLogger.Sugar().Error("Failed to initialize exporters", "err", err)
		log.Fatal(err)
	}

	errChan := make(chan error, 2)

	go func() {
		if err := server.RunGRPCServer(ctx, zapLogger, grpcPort, reg); err != nil {
			zapLogger.Error("gRPC server error", zap.Error(err))
			log.Fatal(err)
		}
	}()

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

	select {
	case <-ctx.Done():
		zapLogger.Info("Shutting down server")
	case err := <-errChan:
		zapLogger.Sugar().Error("Server exited with error", "err", err)
	}
}
