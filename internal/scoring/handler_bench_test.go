package scoring

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"esgbook-software-engineer-technical-test-2024/middleware"
)

func BenchmarkCalculateScoreHandler(b *testing.B) {
	gin.SetMode(gin.TestMode)

	logger, _ := middleware.InitializeLogger()

	handler := &Handler{
		Ctx:            nil, // The handler uses the request's context.
		Logger:         logger,
		ConfigFileName: "score_1.yaml",
	}

	r := gin.New()
	r.GET("/run-scores", handler.CalculateScoreHandler)

	req := httptest.NewRequest("GET", "/run-scores", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
	}
}
