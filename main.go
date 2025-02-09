package main

import (
	"context"
	"os/signal"
	"sync/atomic"

	"log"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc/reflection"

	"github.com/gin-gonic/gin"

	"os"
	"time"

	"go.opentelemetry.io/otel"

	s "esgbook-software-engineer-technical-test-2024/internal/scoring"
	m "esgbook-software-engineer-technical-test-2024/middleware"

	"esgbook-software-engineer-technical-test-2024/grpc-server/logger"
	"esgbook-software-engineer-technical-test-2024/grpc-server/protocol/grpc"
	"esgbook-software-engineer-technical-test-2024/grpc-server/protocol/grpc/middleware/grpctracing"
	"esgbook-software-engineer-technical-test-2024/middleware"
)

const timeout = 10

const file = "score_1.yaml"

// --- Server components

// isReady is used for kube liveness probes, it's only latched to true once
// the gRPC server is ready to handle requests
var isReady atomic.Value

func initializeLogger() error {
	return logger.Init(
		zap.DebugLevel,
		zap.String("service", "example"),
		zap.String("version", "v42.0.69"),
		zap.Strings("maintainers", []string{"@fc", "@FACorreiaa"}),
	)
}

func RunHTTPServer(ctx context.Context, zapLogger *zap.Logger, port string) error {
	router := gin.New()
	router.Use(otelgin.Middleware("score-app"))
	router.Use(gin.Recovery())
	router.Use(logger.ZapLoggingMiddleware(zapLogger))

	h := s.Handler{
		Ctx:            ctx,
		Logger:         zapLogger,
		ConfigFileName: file,
	}

	router.GET("/run-scores", h.CalculateScoreHandler)
	router.GET("/health", s.HealthCheckHandler)

	// 4. Start serving in a blocking manner.
	zapLogger.Info("Starting Gin service on :" + port)
	if err := router.Run(":" + port); err != nil {
		zapLogger.Sugar().Error("Failed to start Gin server", "err", err)
		return err
	}
	return nil
}

func RunGRPCServer(ctx context.Context, zapLogger *zap.Logger, port string, reg *prometheus.Registry) error {

	// Initialize OpenTelemetry trace provider with options as needed
	exp, err := grpctracing.NewOTLPExporter(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to configure OpenTelemetry trace provider")
	}
	tp := grpctracing.NewTraceProvider(exp)
	defer func() { _ = tp.Shutdown(ctx) }()

	otel.SetTracerProvider(tp)
	//tracer = tp.Tracer("myapp")

	// Bootstrap the gRPC server
	server, listener, err := grpc.BootstrapServer(port, zapLogger, reg, tp)
	if err != nil {
		return errors.Wrap(err, "failed to configure gRPC server")
	}

	// Register your services
	//cpb.RegisterCustomerServer(server, container.CustomerService)
	//upb.RegisterAuthServer(server, container.AuthService)
	//ccpb.RegisterCalculatorServer(server, container.CalculatorService)
	//apb.RegisterActivityServer(server, container.ServiceActivity)
	//wpb.RegisterWorkoutServer(server, container.WorkoutService)
	//mpb.RegisterUserMeasurementsServer(server, container.MeasurementService)
	//
	//mlpb.RegisterMealPlanServer(server, container.MealServices.MealPlanService)
	//mlpb.RegisterDietPreferenceServiceServer(server, container.MealServices.DietPreferenceService)
	//mlpb.RegisterFoodLogServiceServer(server, container.MealServices.FoodLogService)
	//mlpb.RegisterIngredientsServer(server, container.MealServices.IngredientService)
	//mlpb.RegisterTrackMealProgressServer(server, container.MealServices.TrackMealProgressService)
	//mlpb.RegisterGoalRecommendationServer(server, container.MealServices.GoalRecommendationService)
	//mlpb.RegisterMealReminderServer(server, container.MealServices.MealReminderService)

	// meal services
	//mealServices := []mlpb.MealServer{
	//	container.MealServices.MealPlanService,
	//	container.MealServices.MealService,
	//	container.MealServices.DietPreferenceService,
	//	container.MealServices.FoodLogService,
	//	container.MealServices.IngredientService,
	//	container.MealServices.TrackMealProgressService,
	//	container.MealServices.GoalRecommendationService,
	//	container.MealServices.MealReminderService,
	//}
	//
	//for _, service := range mealServices {
	//	mlpb.RegisterMealServer(server, service)
	//}

	//mlpb.RegisterMealServer(server, container.MealServices.MealService)
	// Enable gRPC reflection for easier debugging
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

	// Start serving
	zapLogger.Info("gRPC server starting", zap.String("port", port))
	if err = server.Serve(listener); err != nil {
		return errors.Wrap(err, "gRPC server failed to serve")
	}

	isReady.Store(true)
	zapLogger.Info("running grpc server", zap.String("port", port))

	return server.Serve(listener)
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

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Minute)
	defer cancel()

	// Initialize the multi-exporter (Tempo and Jaeger) and set the global tracer provider.
	if err := m.InitExporters(ctx); err != nil {
		zapLogger.Sugar().Error("Failed to initialize exporters", "err", err)
		log.Fatal(err)
	}

	// Start your server and Prometheus concurrently.
	errChan := make(chan error, 2)

	go func() {
		if err := RunHTTPServer(ctx, zapLogger, serverPort); err != nil {
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
		if err := RunGRPCServer(ctx, zapLogger, grpcPort, reg); err != nil {
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
