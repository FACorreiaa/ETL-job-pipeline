package scoring

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "esgbook-software-engineer-technical-test-2024/protos/modules/scoring/generated"
	"esgbook-software-engineer-technical-test-2024/protos/protocol/grpc/middleware/grpcrequest"
)

type GrpcScoringServer struct {
	pb.UnimplementedScoringServiceServer
	Logger         *zap.Logger
	ConfigFileName string
}

func (s *GrpcScoringServer) CalculateScores(ctx context.Context, req *pb.CalculateRequest) (*pb.CalculateResponse, error) {
	// Call your existing business logic:
	tracer := otel.Tracer("score-app")
	_, span := tracer.Start(ctx, "CalculateScoreHTTP")
	defer span.End()
	requestID, ok := ctx.Value(grpcrequest.RequestIDKey{}).(string)
	if !ok {
		requestID = req.GetRequest().GetRequestId() // fallback
	}

	_, scoredResults, err := CalculateScore(ctx, s.Logger, s.ConfigFileName, NewDataLoaderService(NewLoaderRegistry()))
	if err != nil {
		s.Logger.Error("Failed to calculate scores", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "Failed to calculate scores: %v", err)
	}

	var scores []*pb.CompanyScore
	for _, sr := range scoredResults {
		companyScore := &pb.CompanyScore{
			CompanyId: sr.Key.CompanyID,
			Year:      int32(sr.Key.Year),
			Metrics:   make(map[string]float64),
		}
		for metricName, metricVal := range sr.Metrics {
			companyScore.Metrics[metricName] = metricVal
		}
		scores = append(scores, companyScore)
	}

	span.SetAttributes(
		attribute.String("request.id", requestID),
		attribute.String("request.details", req.String()),
	)

	return &pb.CalculateResponse{
		Success: true,
		Message: "Score calculation successful",
		Scores:  scores,
		Response: &pb.BaseResponse{
			Upstream:  "scoring-service",
			RequestId: requestID,
			Status:    "OK",
		},
	}, nil
}
