package scoring

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.uber.org/zap"
)

type Handler struct {
	Ctx            context.Context
	Logger         *zap.Logger
	ConfigFileName string
}

// CalculateScoreHandler Calculate scores and print in csv format
func (h *Handler) CalculateScoreHandler(c *gin.Context) {
	ctx := c.Request.Context()

	tracer := otel.Tracer("score-app")
	_, span := tracer.Start(ctx, "CalculateScoreHTTP")
	defer span.End()

	h.Logger.Info("Calculating score")

	lr := NewLoaderRegistry()
	dataService := NewDataLoaderService(lr)

	scoreConfig, scoredResults, err := CalculateScore(ctx, h.Logger, h.ConfigFileName, dataService)
	if err != nil {
		h.Logger.Info(fmt.Sprintf("Error calculating score: %s", err.Error()))
		c.String(http.StatusInternalServerError, "Error: %v", err)
		return
	}
	//span.SetAttributes(
	//	attribute.String("request.id", requestID),
	//)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", `attachment; filename="scores.csv"`)

	csvWriter := csv.NewWriter(c.Writer)
	defer csvWriter.Flush()

	header := []string{"company", "year"}
	for _, metric := range scoreConfig.Metrics {
		header = append(header, metric.Name)
	}
	err = csvWriter.Write(header)
	if err != nil {
		h.Logger.Info(fmt.Sprintf("Error writing header: %s", err.Error()))
		c.String(http.StatusInternalServerError, "Failed to write CSV header")
		return
	}

	for _, sr := range scoredResults {
		row := []string{
			sr.Key.CompanyID,
			strconv.Itoa(sr.Key.Year),
		}
		for _, metric := range scoreConfig.Metrics {
			if val, ok := sr.Metrics[metric.Name]; ok {
				row = append(row, fmt.Sprintf("%.2f", val))
			} else {
				row = append(row, "") // or "NULL"
			}
		}

		if err = csvWriter.Write(row); err != nil {
			h.Logger.Info(fmt.Sprintf("Error writing row: %s", err.Error()))
			c.String(http.StatusInternalServerError, "Failed to write CSV row")
			return
		}
	}
}

func HealthCheckHandler(c *gin.Context) {
	if err := isServiceHealthy(); err != nil {
		// If the service is NOT healthy:
		c.Status(http.StatusInternalServerError)
		return
	}
	// If the service is OK:
	c.Status(http.StatusOK)
}

func isServiceHealthy() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	select {
	case <-time.After(2 * time.Second):
		return nil
	case <-ctx.Done():
		return fmt.Errorf("timeout while checking service health")
	}
}
