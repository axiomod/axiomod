package persistence

import (
	"context"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/platform/ent"
	entexample "github.com/axiomod/axiomod/platform/ent/example"

	"go.uber.org/zap"
)

// ExampleEntRepository implements the ExampleRepository interface with the
// Ent ORM (the framework default, database.orm: "ent"). The schema lives in
// platform/ent/schema; regenerate with `go generate ./platform/ent/...`.
type ExampleEntRepository struct {
	client *ent.Client
	logger *observability.Logger
}

// NewExampleEntRepository creates a new ExampleEntRepository.
func NewExampleEntRepository(client *ent.Client, logger *observability.Logger) *ExampleEntRepository {
	return &ExampleEntRepository{
		client: client,
		logger: logger,
	}
}

// Migrate creates or updates the underlying schema. Call once at startup.
func (r *ExampleEntRepository) Migrate(ctx context.Context) error {
	return r.client.Schema.Create(ctx)
}

// Create creates a new Example entity
func (r *ExampleEntRepository) Create(ctx context.Context, example *entity.Example) error {
	_, err := r.client.Example.Create().
		SetID(example.ID).
		SetName(example.Name).
		SetDescription(example.Description).
		SetValueType(example.Value.Type).
		SetValueCount(example.Value.Count).
		SetValueTags(example.Value.Tags).
		SetCreatedAt(example.CreatedAt).
		SetUpdatedAt(example.UpdatedAt).
		Save(ctx)
	if err != nil {
		if ent.IsConstraintError(err) {
			return repository.ErrDuplicateID
		}
		r.logger.Error("Failed to create example", zap.Error(err), zap.String("id", example.ID))
		return err
	}
	return nil
}

// GetByID retrieves an Example entity by ID
func (r *ExampleEntRepository) GetByID(ctx context.Context, id string) (*entity.Example, error) {
	row, err := r.client.Example.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, repository.ErrExampleNotFound
		}
		r.logger.Error("Failed to get example", zap.Error(err), zap.String("id", id))
		return nil, err
	}
	return entFromRow(row), nil
}

// Update updates an existing Example entity
func (r *ExampleEntRepository) Update(ctx context.Context, example *entity.Example) error {
	_, err := r.client.Example.UpdateOneID(example.ID).
		SetName(example.Name).
		SetDescription(example.Description).
		SetValueType(example.Value.Type).
		SetValueCount(example.Value.Count).
		SetValueTags(example.Value.Tags).
		SetUpdatedAt(example.UpdatedAt).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return repository.ErrExampleNotFound
		}
		r.logger.Error("Failed to update example", zap.Error(err), zap.String("id", example.ID))
		return err
	}
	return nil
}

// Delete deletes an Example entity by ID
func (r *ExampleEntRepository) Delete(ctx context.Context, id string) error {
	err := r.client.Example.DeleteOneID(id).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return repository.ErrExampleNotFound
		}
		r.logger.Error("Failed to delete example", zap.Error(err), zap.String("id", id))
		return err
	}
	return nil
}

// List retrieves Example entities matching the filter. Name and value-type
// filters run in SQL; the tag filter is applied in memory before pagination
// because portable JSON-array membership differs across dialects.
func (r *ExampleEntRepository) List(ctx context.Context, filter repository.ExampleFilter) ([]*entity.Example, error) {
	query := r.client.Example.Query().Order(ent.Asc(entexample.FieldCreatedAt))

	if filter.Name != "" {
		query = query.Where(entexample.Name(filter.Name))
	}
	if filter.ValueType != "" {
		query = query.Where(entexample.ValueType(filter.ValueType))
	}

	rows, err := query.All(ctx)
	if err != nil {
		r.logger.Error("Failed to list examples", zap.Error(err))
		return nil, err
	}

	result := make([]*entity.Example, 0, len(rows))
	for _, row := range rows {
		example := entFromRow(row)
		if filter.Tag != "" && !example.Value.HasTag(filter.Tag) {
			continue
		}
		result = append(result, example)
	}

	// Apply pagination with the same semantics as the other implementations.
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

// entFromRow maps a generated Ent row back onto the domain entity.
func entFromRow(row *ent.Example) *entity.Example {
	tags := make([]string, len(row.ValueTags))
	copy(tags, row.ValueTags)
	return &entity.Example{
		ID:          row.ID,
		Name:        row.Name,
		Description: row.Description,
		Value: entity.ExampleValue{
			Type:  row.ValueType,
			Count: row.ValueCount,
			Tags:  tags,
		},
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}
