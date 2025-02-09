package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// LoggingMiddleware create standard logger to inject as middleware
//func LoggingMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
//	return func(next http.Handler) http.Handler {
//		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//			start := time.Now()
//
//			next.ServeHTTP(w, r)
//
//			logger.Info("handled request",
//				slog.String("method", r.Method),
//				slog.String("path", r.URL.Path),
//				slog.String("remote_addr", r.RemoteAddr),
//				slog.Duration("duration", time.Since(start)),
//			)
//		})
//	}
//}
//
//func InitLogger() *slog.Logger {
//	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
//		Level:     slog.LevelInfo,
//		AddSource: true,
//		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
//			// Example: rename the "time" key and format the timestamp differently
//			if a.Key == slog.TimeKey && a.Value.Kind() == slog.KindTime {
//				t := a.Value.Time()
//				return slog.Attr{
//					Key:   "timestamp",
//					Value: slog.StringValue(t.Format(time.RFC3339Nano)),
//				}
//			}
//
//			// Example: remove the "source" attribute if you don't want it
//			if a.Key == slog.SourceKey {
//				// Return slog.Attr{} means "drop this attr"
//				return slog.Attr{}
//			}
//
//			// Otherwise, leave all other attributes as-is.
//			return a
//		},
//	}))
//	slog.SetDefault(logger)
//	return logger
//}

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
