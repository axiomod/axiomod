package persistence

import (
	"context"
	"sync"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
)

// ExampleMemoryRepository implements the ExampleRepository interface with in-memory storage
type ExampleMemoryRepository struct {
	examples map[string]*entity.Example
	mu       sync.RWMutex
}

// NewExampleMemoryRepository creates a new ExampleMemoryRepository
func NewExampleMemoryRepository() *ExampleMemoryRepository {
	return &ExampleMemoryRepository{
		examples: make(map[string]*entity.Example),
	}
}

// Create creates a new Example entity
func (r *ExampleMemoryRepository) Create(ctx context.Context, example *entity.Example) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if example with this ID already exists
	if _, exists := r.examples[example.ID]; exists {
		return repository.ErrDuplicateID
	}

	// Store a copy of the example
	r.examples[example.ID] = cloneExample(example)
	return nil
}

// GetByID retrieves an Example entity by ID
func (r *ExampleMemoryRepository) GetByID(ctx context.Context, id string) (*entity.Example, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Check if example exists
	example, exists := r.examples[id]
	if !exists {
		return nil, repository.ErrExampleNotFound
	}

	// Return a copy of the example
	return cloneExample(example), nil
}

// Update updates an existing Example entity
func (r *ExampleMemoryRepository) Update(ctx context.Context, example *entity.Example) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if example exists
	if _, exists := r.examples[example.ID]; !exists {
		return repository.ErrExampleNotFound
	}

	// Update the example
	r.examples[example.ID] = cloneExample(example)
	return nil
}

// Delete deletes an Example entity by ID
func (r *ExampleMemoryRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if example exists
	if _, exists := r.examples[id]; !exists {
		return repository.ErrExampleNotFound
	}

	// Delete the example
	delete(r.examples, id)
	return nil
}

// List retrieves all Example entities with optional filtering
func (r *ExampleMemoryRepository) List(ctx context.Context, filter repository.ExampleFilter) ([]*entity.Example, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*entity.Example

	// Apply filters
	for _, example := range r.examples {
		if matchesFilter(example, filter) {
			result = append(result, cloneExample(example))
		}
	}

	// Apply pagination
	if filter.Limit > 0 {
		offset := filter.Offset
		if offset >= len(result) {
			return []*entity.Example{}, nil
		}

		end := offset + filter.Limit
		if end > len(result) {
			end = len(result)
		}

		result = result[offset:end]
	}

	return result, nil
}

// matchesFilter checks if an example matches the filter criteria
func matchesFilter(example *entity.Example, filter repository.ExampleFilter) bool {
	// Filter by name
	if filter.Name != "" && example.Name != filter.Name {
		return false
	}

	// Filter by value type
	if filter.ValueType != "" && example.Value.Type != filter.ValueType {
		return false
	}

	// Filter by tag
	if filter.Tag != "" && !example.Value.HasTag(filter.Tag) {
		return false
	}

	return true
}

// cloneExample creates a deep copy of an Example entity
func cloneExample(example *entity.Example) *entity.Example {
	// Clone tags
	tags := make([]string, len(example.Value.Tags))
	copy(tags, example.Value.Tags)

	// Clone value
	value := entity.ExampleValue{
		Type:  example.Value.Type,
		Count: example.Value.Count,
		Tags:  tags,
	}

	// Clone example
	return &entity.Example{
		ID:          example.ID,
		Name:        example.Name,
		Description: example.Description,
		Value:       value,
		CreatedAt:   example.CreatedAt,
		UpdatedAt:   example.UpdatedAt,
	}
}
