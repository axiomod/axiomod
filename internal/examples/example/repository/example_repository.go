package repository

import (
	"context"

	"axiomod/internal/examples/example/entity"
)

// ExampleRepository defines the interface for Example entity persistence
type ExampleRepository interface {
	// Create creates a new Example entity
	Create(ctx context.Context, example *entity.Example) error

	// GetByID retrieves an Example entity by ID
	GetByID(ctx context.Context, id string) (*entity.Example, error)

	// Update updates an existing Example entity
	Update(ctx context.Context, example *entity.Example) error

	// Delete deletes an Example entity by ID
	Delete(ctx context.Context, id string) error

	// List retrieves all Example entities with optional filtering
	List(ctx context.Context, filter ExampleFilter) ([]*entity.Example, error)
}

// ExampleFilter defines filtering options for listing Example entities
type ExampleFilter struct {
	Name      string
	ValueType string
	Tag       string
	Limit     int
	Offset    int
}

// Repository errors
var (
	ErrExampleNotFound = entity.NewDomainError("repository.example_not_found", "Example not found")
	ErrDuplicateID     = entity.NewDomainError("repository.duplicate_id", "Example with this ID already exists")
)
