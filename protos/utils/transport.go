package utils

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var Transport *TransportUtils

type TransportUtils struct {
	Logger        *zap.Logger
	Prometheus    *prometheus.Registry
	TraceProvider trace.TracerProvider
}
