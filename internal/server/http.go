package server

import (
	"context"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/zap"

	s "esgbook-software-engineer-technical-test-2024/internal/scoring"
	"esgbook-software-engineer-technical-test-2024/middleware"
)

const file = "score_1.yaml"

func RunHTTPServer(ctx context.Context, zapLogger *zap.Logger, port string) error {
	router := gin.New()
	router.Use(otelgin.Middleware("score-app"))
	router.Use(gin.Recovery())
	router.Use(middleware.ZapLoggingMiddleware(zapLogger))

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
