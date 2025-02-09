package logger

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Log      *zap.Logger
	onceInit sync.Once
)

func Init(level zapcore.Level, meta ...zap.Field) error {
	onceInit.Do(func() {
		instance := zap.Must(configure(level).Build())

		// attach the additional meta fields to the logger before committing to global instance
		instance = instance.With(meta...)
		instance = instance.With(zap.String("line", "42"))

		Log = zap.New(instance.Core(), zap.AddCaller())
	})

	if Log == nil {
		return errors.New("logger not initialized")
	}

	return nil
}

func configure(level zapcore.Level) zap.Config {
	encoder := zap.NewProductionEncoderConfig()
	encoder.TimeKey = "timestamp"
	encoder.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder.EncodeLevel = zapcore.CapitalColorLevelEncoder
	encoder.EncodeCaller = zapcore.ShortCallerEncoder
	encoder.EncodeDuration = zapcore.SecondsDurationEncoder
	encoder.EncodeName = zapcore.FullNameEncoder
	encoder.CallerKey = "caller"
	return zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Encoding:          "console",
		EncoderConfig:     encoder,
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
	}
}

// ZapLoggingMiddleware logs incoming requests using a zap.Logger
func ZapLoggingMiddleware(zapLogger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Stop timer
		latency := time.Since(start)
		status := c.Writer.Status()

		// Log the request details:
		zapLogger.Info("HTTP request",
			zap.Int("status", status),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("latency", latency),
			zap.String("clientIP", c.ClientIP()),
		)
	}
}
