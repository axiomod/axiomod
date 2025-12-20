package usecase

import (
	"context"

	"axiomod/internal/examples/example/entity"
	"axiomod/internal/examples/example/repository"
)

// CreateExampleInput represents the input for creating an example
type CreateExampleInput struct {
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	ValueType   string   `json:"valueType" validate:"required"`
	Count       int      `json:"count" validate:"gte=0"`
	Tags        []string `json:"tags"`
}

// CreateExampleOutput represents the output of creating an example
type CreateExampleOutput struct {
	ID string `json:"id"`
}

// CreateExampleUseCase defines the use case for creating an example
type CreateExampleUseCase struct {
	repo repository.ExampleRepository
}

// NewCreateExampleUseCase creates a new CreateExampleUseCase
func NewCreateExampleUseCase(repo repository.ExampleRepository) *CreateExampleUseCase {
	return &CreateExampleUseCase{
		repo: repo,
	}
}

// Execute executes the use case
func (uc *CreateExampleUseCase) Execute(ctx context.Context, input CreateExampleInput) (*CreateExampleOutput, error) {
	// Create value object
	value := entity.NewExampleValue(input.ValueType, input.Count, input.Tags)

	// Create entity
	example := entity.NewExample(input.Name, input.Description, value)

	// Validate entity
	if err := example.Validate(); err != nil {
		return nil, err
	}

	// Save to repository
	if err := uc.repo.Create(ctx, example); err != nil {
		return nil, err
	}

	return &CreateExampleOutput{
		ID: example.ID,
	}, nil
}
