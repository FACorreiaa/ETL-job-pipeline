package server

import (
	"context"
	"os"
	"os/signal"
	"sync/atomic"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"

	pb "esgbook-software-engineer-technical-test-2024/protos/modules/scoring/generated"
	"esgbook-software-engineer-technical-test-2024/protos/protocol/grpc"
	"esgbook-software-engineer-technical-test-2024/protos/protocol/grpc/middleware/grpctracing"
)

// isReady is used for liveness probes in Kubernetes
var isReady atomic.Value

func RunGRPCServer(ctx context.Context, zapLogger *zap.Logger, port string, reg *prometheus.Registry) error {
	// Initialize OpenTelemetry trace provider
	exp, err := grpctracing.NewOTLPExporter(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to configure OpenTelemetry trace provider")
	}
	tp := grpctracing.NewTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()
	otel.SetTracerProvider(tp)

	b := NewBroker(zapLogger, file, reg)
	zapLogger.Info("Broker initialized")

	zapLogger.Info("Attempting to start gRPC server on", zap.String("port", port))
	server, listener, err := grpc.BootstrapServer(port, zapLogger, reg, tp)
	if err != nil {
		return errors.Wrap(err, "failed to configure gRPC server")
	}
	zapLogger.Info("gRPC server successfully configured", zap.String("port", port))

	// Register gRPC services using Broker
	pb.RegisterScoringServiceServer(server, b.GetScoringService())

	// Enable reflection for debugging
	reflection.Register(server)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			zapLogger.Warn("shutting down grpc server")
			server.GracefulStop()
			<-ctx.Done()
		}
	}()

	isReady.Store(true)
	zapLogger.Info("gRPC server starting", zap.String("bindAddress", "0.0.0.0:"+port))
	err = server.Serve(listener)
	if err != nil {
		zapLogger.Error("gRPC server exited with error", zap.Error(err))
		return errors.Wrap(err, "gRPC server failed to serve")
	}
	return nil
}
