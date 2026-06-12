package usecase

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
)

// UpdateExampleInput represents the input for updating an example
type UpdateExampleInput struct {
	ID          string   `json:"id" validate:"required"`
	Name        string   `json:"name" validate:"required"`
	Description string   `json:"description" validate:"required"`
	ValueType   string   `json:"valueType" validate:"required"`
	Count       int      `json:"count" validate:"gte=0"`
	Tags        []string `json:"tags"`
}

// UpdateExampleOutput represents the output of updating an example
type UpdateExampleOutput struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ValueType   string   `json:"valueType"`
	Count       int      `json:"count"`
	Tags        []string `json:"tags"`
	UpdatedAt   string   `json:"updatedAt"`
}

// UpdateExampleUseCase defines the use case for updating an example
type UpdateExampleUseCase struct {
	repo repository.ExampleRepository
}

// NewUpdateExampleUseCase creates a new UpdateExampleUseCase
func NewUpdateExampleUseCase(repo repository.ExampleRepository) *UpdateExampleUseCase {
	return &UpdateExampleUseCase{
		repo: repo,
	}
}

// Execute executes the use case
func (uc *UpdateExampleUseCase) Execute(ctx context.Context, input UpdateExampleInput) (*UpdateExampleOutput, error) {
	// Load the current entity
	example, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Apply the changes through the entity's behavior
	value := entity.NewExampleValue(input.ValueType, input.Count, input.Tags)
	example.Update(input.Name, input.Description, value)
	if err := example.Validate(); err != nil {
		return nil, err
	}

	// Persist
	if err := uc.repo.Update(ctx, example); err != nil {
		return nil, err
	}

	return &UpdateExampleOutput{
		ID:          example.ID,
		Name:        example.Name,
		Description: example.Description,
		ValueType:   example.Value.Type,
		Count:       example.Value.Count,
		Tags:        example.Value.Tags,
		UpdatedAt:   example.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
