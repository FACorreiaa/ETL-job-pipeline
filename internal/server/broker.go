package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"

	"esgbook-software-engineer-technical-test-2024/internal/scoring"
	pb "esgbook-software-engineer-technical-test-2024/protos/modules/scoring/generated"
)

// Broker manages the gRPC service lifecycle
type Broker struct {
	Logger         *zap.Logger
	ConfigFileName string
	Registry       *prometheus.Registry
}

// NewBroker initializes a new Broker
func NewBroker(logger *zap.Logger, configFileName string, registry *prometheus.Registry) *Broker {
	return &Broker{
		Logger:         logger,
		ConfigFileName: configFileName,
		Registry:       registry,
	}
}

// GetScoringService returns an instance of the gRPC scoring service
func (b *Broker) GetScoringService() pb.ScoringServiceServer {
	return &scoring.GrpcScoringServer{
		Logger:         b.Logger,
		ConfigFileName: b.ConfigFileName,
	}
}
