package grpc

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/usecase"
	"github.com/axiomod/axiomod/framework/observability"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ExampleGRPCService implements the gRPC service for the Example entity
type ExampleGRPCService struct {
	createUseCase *usecase.CreateExampleUseCase
	getUseCase    *usecase.GetExampleUseCase
	logger        *observability.Logger
	UnimplementedExampleServiceServer
}

// NewExampleGRPCService creates a new ExampleGRPCService
func NewExampleGRPCService(
	createUseCase *usecase.CreateExampleUseCase,
	getUseCase *usecase.GetExampleUseCase,
	logger *observability.Logger,
) *ExampleGRPCService {
	return &ExampleGRPCService{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		logger:        logger,
	}
}

// CreateExample handles the creation of a new Example via gRPC
func (s *ExampleGRPCService) CreateExample(ctx context.Context, req *CreateExampleRequest) (*CreateExampleResponse, error) {
	// Map request to use case input
	input := usecase.CreateExampleInput{
		Name:        req.Name,
		Description: req.Description,
		ValueType:   req.ValueType,
		Count:       int(req.Count),
		Tags:        req.Tags,
	}

	// Execute use case
	output, err := s.createUseCase.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Failed to create example", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return response
	return &CreateExampleResponse{
		Id: output.ID,
	}, nil
}

// GetExample handles the retrieval of an Example by ID via gRPC
func (s *ExampleGRPCService) GetExample(ctx context.Context, req *GetExampleRequest) (*GetExampleResponse, error) {
	// Execute use case
	output, err := s.getUseCase.Execute(ctx, usecase.GetExampleInput{ID: req.Id})
	if err != nil {
		s.logger.Error("Failed to get example", zap.Error(err), zap.String("id", req.Id))
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Return response
	return &GetExampleResponse{
		Id:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		ValueType:   output.ValueType,
		Count:       int32(output.Count),
		Tags:        output.Tags,
		CreatedAt:   output.CreatedAt,
		UpdatedAt:   output.UpdatedAt,
	}, nil
}

// The request/response message types and the ExampleServiceServer interface
// are generated from example.proto (see example.pb.go and example_grpc.pb.go).
// Regenerate with:
//
//	protoc --go_out=. --go_opt=paths=source_relative \
//	       --go-grpc_out=. --go-grpc_opt=paths=source_relative example.proto
