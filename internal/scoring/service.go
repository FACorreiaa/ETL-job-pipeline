package scoring

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	c "esgbook-software-engineer-technical-test-2024/pkg/config"
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

func (s *GrpcScoringServer) CalculateScoresStream(req *pb.CalculateRequest, stream pb.ScoringService_CalculateScoresStreamServer) error {
	ctx := stream.Context()

	// Start a trace span.
	tracer := otel.Tracer("score-app")
	_, span := tracer.Start(ctx, "CalculateScoresStream")
	defer span.End()

	// Extract request ID either from context or request fallback.
	requestID, ok := ctx.Value(grpcrequest.RequestIDKey{}).(string)
	if !ok {
		requestID = req.GetRequest().GetRequestId()
	}
	span.SetAttributes(attribute.String("request.id", requestID))
	s.Logger.Info("Starting streaming score calculation", zap.String("request_id", requestID))

	scoreConfig, err := c.InitScoreConfig(s.ConfigFileName)
	if err != nil {
		s.Logger.Error("Failed to initialize score config", zap.Error(err))
		return status.Errorf(codes.Internal, "failed to initialize score config: %v", err)
	}

	graph, inDegree := buildDependencyGraph(s.Logger, scoreConfig)
	topoOrder, err := topologicalSort(s.Logger, scoreConfig, graph, inDegree)
	if err != nil {
		s.Logger.Error("Failed topological sort", zap.Error(err))
		return status.Errorf(codes.Internal, "failed topological sort: %v", err)
	}

	metricMap := BuildMetricMap(scoreConfig)

	dataService := NewDataLoaderService(NewLoaderRegistry())
	combined, err := dataService.LoadAllData(ctx, Dir)
	if err != nil {
		s.Logger.Error("Failed to load data from folder", zap.Error(err))
		return status.Errorf(codes.Internal, "failed to load data from folder: %v", err)
	}
	datasets := make(map[string]map[CompanyYearKey]map[string]float64)
	for logicalName, csvKey := range DatasetKeys {
		data, ok := combined[csvKey]
		if !ok {
			s.Logger.Error("Missing dataset", zap.String("key", csvKey))
			return status.Errorf(codes.Internal, "missing dataset for key %q", csvKey)
		}
		datasets[logicalName] = data
	}
	allKeys := getAllDataCompanyKeys(datasets)

	scoreCh, err := StreamScores(ctx, s.Logger, allKeys, topoOrder, metricMap, datasets, NumWorkers)
	if err != nil {
		s.Logger.Error("Failed to stream scores", zap.Error(err))
		return status.Errorf(codes.Internal, "failed to stream scores: %v", err)
	}

	for score := range scoreCh {
		companyScore := &pb.CompanyScore{
			CompanyId: score.Key.CompanyID,
			Year:      int32(score.Key.Year),
			Metrics:   make(map[string]float64),
		}
		for metricName, metricVal := range score.Metrics {
			companyScore.Metrics[metricName] = metricVal
		}
		if err := stream.Send(companyScore); err != nil {
			s.Logger.Error("Failed to send score over stream", zap.Error(err))
			return status.Errorf(codes.Internal, "failed to send score: %v", err)
		}
	}
	s.Logger.Info("Finished streaming scores", zap.String("request_id", requestID))
	return nil
}
