package service

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
)

// ExampleDomainService provides domain-specific business logic for Example entities
type ExampleDomainService struct {
	repo repository.ExampleRepository
}

// NewExampleDomainService creates a new ExampleDomainService
func NewExampleDomainService(repo repository.ExampleRepository) *ExampleDomainService {
	return &ExampleDomainService{
		repo: repo,
	}
}

// ProcessExample performs domain-specific processing on an Example entity
func (s *ExampleDomainService) ProcessExample(ctx context.Context, example *entity.Example) error {
	// Example of domain-specific business logic that spans multiple entities
	// or requires complex processing beyond simple CRUD operations

	// For example, we might:
	// 1. Validate business rules that span multiple entities
	// 2. Perform calculations or transformations
	// 3. Check for conflicts or dependencies

	// Add a special tag based on the value type
	if example.Value.Type == "premium" && !example.Value.HasTag("premium") {
		example.Value.AddTag("premium")
		return s.repo.Update(ctx, example)
	}

	return nil
}

// FindRelatedExamples finds examples related to the given example
func (s *ExampleDomainService) FindRelatedExamples(ctx context.Context, exampleID string) ([]*entity.Example, error) {
	// Get the example
	example, err := s.repo.GetByID(ctx, exampleID)
	if err != nil {
		return nil, err
	}

	// Find related examples by tags
	if len(example.Value.Tags) > 0 {
		filter := repository.ExampleFilter{
			Tag: example.Value.Tags[0], // Use the first tag for simplicity
		}
		return s.repo.List(ctx, filter)
	}

	return []*entity.Example{}, nil
}
