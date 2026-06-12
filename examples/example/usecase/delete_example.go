package usecase

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/repository"
)

// DeleteExampleInput represents the input for deleting an example
type DeleteExampleInput struct {
	ID string `json:"id" validate:"required"`
}

// DeleteExampleOutput represents the output of deleting an example
type DeleteExampleOutput struct {
	ID      string `json:"id"`
	Deleted bool   `json:"deleted"`
}

// DeleteExampleUseCase defines the use case for deleting an example
type DeleteExampleUseCase struct {
	repo repository.ExampleRepository
}

// NewDeleteExampleUseCase creates a new DeleteExampleUseCase
func NewDeleteExampleUseCase(repo repository.ExampleRepository) *DeleteExampleUseCase {
	return &DeleteExampleUseCase{
		repo: repo,
	}
}

// Execute executes the use case
func (uc *DeleteExampleUseCase) Execute(ctx context.Context, input DeleteExampleInput) (*DeleteExampleOutput, error) {
	if err := uc.repo.Delete(ctx, input.ID); err != nil {
		return nil, err
	}

	return &DeleteExampleOutput{
		ID:      input.ID,
		Deleted: true,
	}, nil
}
