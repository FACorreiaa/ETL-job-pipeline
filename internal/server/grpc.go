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

	"esgbook-software-engineer-technical-test-2024/protos/protocol/grpc"
	"esgbook-software-engineer-technical-test-2024/protos/protocol/grpc/middleware/grpctracing"
)

// --- Server components

// isReady is used for kube liveness probes, it's only latched to true once
// the gRPC server is ready to handle requests
var isReady atomic.Value

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
