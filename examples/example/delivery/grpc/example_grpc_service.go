package grpc

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/usecase"
	"github.com/axiomod/axiomod/platform/observability"

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

// Note: In a real implementation, we would have generated gRPC service definitions
// from protobuf files. For this example, we're defining placeholder types.

// UnimplementedExampleServiceServer is a placeholder for the generated gRPC server interface
type UnimplementedExampleServiceServer struct{}

// CreateExampleRequest represents the request for creating an example
type CreateExampleRequest struct {
	Name        string   `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Description string   `protobuf:"bytes,2,opt,name=description,proto3" json:"description,omitempty"`
	ValueType   string   `protobuf:"bytes,3,opt,name=value_type,json=valueType,proto3" json:"value_type,omitempty"`
	Count       int32    `protobuf:"varint,4,opt,name=count,proto3" json:"count,omitempty"`
	Tags        []string `protobuf:"bytes,5,rep,name=tags,proto3" json:"tags,omitempty"`
}

// CreateExampleResponse represents the response for creating an example
type CreateExampleResponse struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

// GetExampleRequest represents the request for getting an example
type GetExampleRequest struct {
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
}

// GetExampleResponse represents the response for getting an example
type GetExampleResponse struct {
	Id          string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name        string   `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Description string   `protobuf:"bytes,3,opt,name=description,proto3" json:"description,omitempty"`
	ValueType   string   `protobuf:"bytes,4,opt,name=value_type,json=valueType,proto3" json:"value_type,omitempty"`
	Count       int32    `protobuf:"varint,5,opt,name=count,proto3" json:"count,omitempty"`
	Tags        []string `protobuf:"bytes,6,rep,name=tags,proto3" json:"tags,omitempty"`
	CreatedAt   string   `protobuf:"bytes,7,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt   string   `protobuf:"bytes,8,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
}
