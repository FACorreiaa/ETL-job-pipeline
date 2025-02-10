package scoring

import (
	"context"
	"testing"

	"esgbook-software-engineer-technical-test-2024/middleware"
	pb "esgbook-software-engineer-technical-test-2024/protos/modules/scoring/generated"
)

func BenchmarkCalculateScoresGRPC(b *testing.B) {
	logger, _ := middleware.InitializeLogger()

	grpcHandler := &GrpcScoringServer{
		Logger:         logger,
		ConfigFileName: "score_1.yaml",
	}

	req := &pb.CalculateRequest{}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := grpcHandler.CalculateScores(ctx, req)
		if err != nil {
			b.Fatalf("gRPC handler failed: %v", err)
		}
	}
}
