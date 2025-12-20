package usecase

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/repository"
)

// GetExampleInput represents the input for getting an example
type GetExampleInput struct {
	ID string `json:"id" validate:"required"`
}

// GetExampleOutput represents the output of getting an example
type GetExampleOutput struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	ValueType   string   `json:"valueType"`
	Count       int      `json:"count"`
	Tags        []string `json:"tags"`
	CreatedAt   string   `json:"createdAt"`
	UpdatedAt   string   `json:"updatedAt"`
}

// GetExampleUseCase defines the use case for getting an example
type GetExampleUseCase struct {
	repo repository.ExampleRepository
}

// NewGetExampleUseCase creates a new GetExampleUseCase
func NewGetExampleUseCase(repo repository.ExampleRepository) *GetExampleUseCase {
	return &GetExampleUseCase{
		repo: repo,
	}
}

// Execute executes the use case
func (uc *GetExampleUseCase) Execute(ctx context.Context, input GetExampleInput) (*GetExampleOutput, error) {
	// Get from repository
	example, err := uc.repo.GetByID(ctx, input.ID)
	if err != nil {
		return nil, err
	}

	// Map to output
	return &GetExampleOutput{
		ID:          example.ID,
		Name:        example.Name,
		Description: example.Description,
		ValueType:   example.Value.Type,
		Count:       example.Value.Count,
		Tags:        example.Value.Tags,
		CreatedAt:   example.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   example.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
