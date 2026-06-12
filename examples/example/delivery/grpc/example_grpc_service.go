package grpc

import (
	"context"
	"errors"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
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
	updateUseCase *usecase.UpdateExampleUseCase
	deleteUseCase *usecase.DeleteExampleUseCase
	listUseCase   *usecase.ListExamplesUseCase
	logger        *observability.Logger
	UnimplementedExampleServiceServer
}

// NewExampleGRPCService creates a new ExampleGRPCService
func NewExampleGRPCService(
	createUseCase *usecase.CreateExampleUseCase,
	getUseCase *usecase.GetExampleUseCase,
	updateUseCase *usecase.UpdateExampleUseCase,
	deleteUseCase *usecase.DeleteExampleUseCase,
	listUseCase *usecase.ListExamplesUseCase,
	logger *observability.Logger,
) *ExampleGRPCService {
	return &ExampleGRPCService{
		createUseCase: createUseCase,
		getUseCase:    getUseCase,
		updateUseCase: updateUseCase,
		deleteUseCase: deleteUseCase,
		listUseCase:   listUseCase,
		logger:        logger,
	}
}

// CreateExample handles the creation of a new Example via gRPC
func (s *ExampleGRPCService) CreateExample(ctx context.Context, req *CreateExampleRequest) (*CreateExampleResponse, error) {
	input := usecase.CreateExampleInput{
		Name:        req.Name,
		Description: req.Description,
		ValueType:   req.ValueType,
		Count:       int(req.Count),
		Tags:        req.Tags,
	}

	output, err := s.createUseCase.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Failed to create example", zap.Error(err))
		return nil, domainStatusError(err)
	}

	return &CreateExampleResponse{
		Id: output.ID,
	}, nil
}

// GetExample handles the retrieval of an Example by ID via gRPC
func (s *ExampleGRPCService) GetExample(ctx context.Context, req *GetExampleRequest) (*GetExampleResponse, error) {
	output, err := s.getUseCase.Execute(ctx, usecase.GetExampleInput{ID: req.Id})
	if err != nil {
		s.logger.Error("Failed to get example", zap.Error(err), zap.String("id", req.Id))
		return nil, domainStatusError(err)
	}

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

// UpdateExample handles updating an existing Example via gRPC
func (s *ExampleGRPCService) UpdateExample(ctx context.Context, req *UpdateExampleRequest) (*UpdateExampleResponse, error) {
	input := usecase.UpdateExampleInput{
		ID:          req.Id,
		Name:        req.Name,
		Description: req.Description,
		ValueType:   req.ValueType,
		Count:       int(req.Count),
		Tags:        req.Tags,
	}

	output, err := s.updateUseCase.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Failed to update example", zap.Error(err), zap.String("id", req.Id))
		return nil, domainStatusError(err)
	}

	return &UpdateExampleResponse{
		Id:          output.ID,
		Name:        output.Name,
		Description: output.Description,
		ValueType:   output.ValueType,
		Count:       int32(output.Count),
		Tags:        output.Tags,
		UpdatedAt:   output.UpdatedAt,
	}, nil
}

// DeleteExample handles deleting an Example by ID via gRPC
func (s *ExampleGRPCService) DeleteExample(ctx context.Context, req *DeleteExampleRequest) (*DeleteExampleResponse, error) {
	output, err := s.deleteUseCase.Execute(ctx, usecase.DeleteExampleInput{ID: req.Id})
	if err != nil {
		s.logger.Error("Failed to delete example", zap.Error(err), zap.String("id", req.Id))
		return nil, domainStatusError(err)
	}

	return &DeleteExampleResponse{
		Id:      output.ID,
		Deleted: output.Deleted,
	}, nil
}

// ListExamples handles listing Examples via gRPC
func (s *ExampleGRPCService) ListExamples(ctx context.Context, req *ListExamplesRequest) (*ListExamplesResponse, error) {
	input := usecase.ListExamplesInput{
		Name:      req.Name,
		ValueType: req.ValueType,
		Tag:       req.Tag,
		Limit:     int(req.Limit),
		Offset:    int(req.Offset),
	}

	output, err := s.listUseCase.Execute(ctx, input)
	if err != nil {
		s.logger.Error("Failed to list examples", zap.Error(err))
		return nil, domainStatusError(err)
	}

	items := make([]*GetExampleResponse, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, &GetExampleResponse{
			Id:          item.ID,
			Name:        item.Name,
			Description: item.Description,
			ValueType:   item.ValueType,
			Count:       int32(item.Count),
			Tags:        item.Tags,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
		})
	}

	return &ListExamplesResponse{
		Items: items,
		Total: int32(output.Total),
	}, nil
}

// domainStatusError maps domain errors onto gRPC status codes: not-found to
// NotFound, validation failures to InvalidArgument, everything else Internal.
func domainStatusError(err error) error {
	var domainErr entity.DomainError
	switch {
	case errors.Is(err, repository.ErrExampleNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.As(err, &domainErr):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}

// The request/response message types and the ExampleServiceServer interface
// are generated from example.proto (see example.pb.go and example_grpc.pb.go).
// Regenerate with protoc or buf using protoc-gen-go and protoc-gen-go-grpc:
//
//	protoc --go_out=. --go_opt=paths=source_relative \
//	       --go-grpc_out=. --go-grpc_opt=paths=source_relative example.proto
