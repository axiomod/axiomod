package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
	"github.com/axiomod/axiomod/platform/observability"

	"go.uber.org/zap"
)

// ExampleEntRepository implements the ExampleRepository interface with Ent ORM
type ExampleEntRepository struct {
	db     *sql.DB
	logger *observability.Logger
}

// NewExampleEntRepository creates a new ExampleEntRepository
func NewExampleEntRepository(db *sql.DB, logger *observability.Logger) *ExampleEntRepository {
	return &ExampleEntRepository{
		db:     db,
		logger: logger,
	}
}

// Create creates a new Example entity
func (r *ExampleEntRepository) Create(ctx context.Context, example *entity.Example) error {
	// In a real implementation, we would use the Ent ORM to create the entity
	// For this example, we'll use a simple SQL query
	query := `
		INSERT INTO examples (id, name, description, value_type, value_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(
		ctx,
		query,
		example.ID,
		example.Name,
		example.Description,
		example.Value.Type,
		example.Value.Count,
		example.CreatedAt,
		example.UpdatedAt,
	)

	if err != nil {
		r.logger.Error("Failed to create example", zap.Error(err))
		return fmt.Errorf("failed to create example: %w", err)
	}

	// Insert tags
	for _, tag := range example.Value.Tags {
		tagQuery := `
			INSERT INTO example_tags (example_id, tag)
			VALUES (?, ?)
		`
		_, err := r.db.ExecContext(ctx, tagQuery, example.ID, tag)
		if err != nil {
			r.logger.Error("Failed to create example tag", zap.Error(err))
			return fmt.Errorf("failed to create example tag: %w", err)
		}
	}

	return nil
}

// GetByID retrieves an Example entity by ID
func (r *ExampleEntRepository) GetByID(ctx context.Context, id string) (*entity.Example, error) {
	// In a real implementation, we would use the Ent ORM to retrieve the entity
	// For this example, we'll use a simple SQL query
	query := `
		SELECT id, name, description, value_type, value_count, created_at, updated_at
		FROM examples
		WHERE id = ?
	`

	row := r.db.QueryRowContext(ctx, query, id)

	var example entity.Example
	var valueType string
	var valueCount int
	var createdAt, updatedAt time.Time

	err := row.Scan(
		&example.ID,
		&example.Name,
		&example.Description,
		&valueType,
		&valueCount,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, repository.ErrExampleNotFound
		}
		r.logger.Error("Failed to get example", zap.Error(err))
		return nil, fmt.Errorf("failed to get example: %w", err)
	}

	example.CreatedAt = createdAt
	example.UpdatedAt = updatedAt

	// Get tags
	tagsQuery := `
		SELECT tag
		FROM example_tags
		WHERE example_id = ?
	`

	rows, err := r.db.QueryContext(ctx, tagsQuery, id)
	if err != nil {
		r.logger.Error("Failed to get example tags", zap.Error(err))
		return nil, fmt.Errorf("failed to get example tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			r.logger.Error("Failed to scan example tag", zap.Error(err))
			return nil, fmt.Errorf("failed to scan example tag: %w", err)
		}
		tags = append(tags, tag)
	}

	example.Value = entity.ExampleValue{
		Type:  valueType,
		Count: valueCount,
		Tags:  tags,
	}

	return &example, nil
}

// Update updates an existing Example entity
func (r *ExampleEntRepository) Update(ctx context.Context, example *entity.Example) error {
	// In a real implementation, we would use the Ent ORM to update the entity
	// For this example, we'll use a simple SQL query
	query := `
		UPDATE examples
		SET name = ?, description = ?, value_type = ?, value_count = ?, updated_at = ?
		WHERE id = ?
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		example.Name,
		example.Description,
		example.Value.Type,
		example.Value.Count,
		time.Now(),
		example.ID,
	)

	if err != nil {
		r.logger.Error("Failed to update example", zap.Error(err))
		return fmt.Errorf("failed to update example: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrExampleNotFound
	}

	// Update tags (delete and re-insert)
	deleteTagsQuery := `
		DELETE FROM example_tags
		WHERE example_id = ?
	`
	_, err = r.db.ExecContext(ctx, deleteTagsQuery, example.ID)
	if err != nil {
		r.logger.Error("Failed to delete example tags", zap.Error(err))
		return fmt.Errorf("failed to delete example tags: %w", err)
	}

	// Insert new tags
	for _, tag := range example.Value.Tags {
		tagQuery := `
			INSERT INTO example_tags (example_id, tag)
			VALUES (?, ?)
		`
		_, err := r.db.ExecContext(ctx, tagQuery, example.ID, tag)
		if err != nil {
			r.logger.Error("Failed to create example tag", zap.Error(err))
			return fmt.Errorf("failed to create example tag: %w", err)
		}
	}

	return nil
}

// Delete deletes an Example entity by ID
func (r *ExampleEntRepository) Delete(ctx context.Context, id string) error {
	// In a real implementation, we would use the Ent ORM to delete the entity
	// For this example, we'll use a simple SQL query

	// Delete tags first (foreign key constraint)
	deleteTagsQuery := `
		DELETE FROM example_tags
		WHERE example_id = ?
	`
	_, err := r.db.ExecContext(ctx, deleteTagsQuery, id)
	if err != nil {
		r.logger.Error("Failed to delete example tags", zap.Error(err))
		return fmt.Errorf("failed to delete example tags: %w", err)
	}

	// Delete example
	query := `
		DELETE FROM examples
		WHERE id = ?
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		r.logger.Error("Failed to delete example", zap.Error(err))
		return fmt.Errorf("failed to delete example: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to get rows affected", zap.Error(err))
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return repository.ErrExampleNotFound
	}

	return nil
}

// List retrieves all Example entities with optional filtering
func (r *ExampleEntRepository) List(ctx context.Context, filter repository.ExampleFilter) ([]*entity.Example, error) {
	// In a real implementation, we would use the Ent ORM to query the entities
	// For this example, we'll use a simple SQL query
	query := `
		SELECT id, name, description, value_type, value_count, created_at, updated_at
		FROM examples
		WHERE 1=1
	`
	var args []interface{}

	// Apply filters
	if filter.Name != "" {
		query += " AND name = ?"
		args = append(args, filter.Name)
	}

	if filter.ValueType != "" {
		query += " AND value_type = ?"
		args = append(args, filter.ValueType)
	}

	// Apply pagination
	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)

		if filter.Offset > 0 {
			query += " OFFSET ?"
			args = append(args, filter.Offset)
		}
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		r.logger.Error("Failed to list examples", zap.Error(err))
		return nil, fmt.Errorf("failed to list examples: %w", err)
	}
	defer rows.Close()

	var examples []*entity.Example
	for rows.Next() {
		var example entity.Example
		var valueType string
		var valueCount int
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&example.ID,
			&example.Name,
			&example.Description,
			&valueType,
			&valueCount,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			r.logger.Error("Failed to scan example", zap.Error(err))
			return nil, fmt.Errorf("failed to scan example: %w", err)
		}

		example.CreatedAt = createdAt
		example.UpdatedAt = updatedAt

		// Get tags for each example
		tagsQuery := `
			SELECT tag
			FROM example_tags
			WHERE example_id = ?
		`

		tagRows, err := r.db.QueryContext(ctx, tagsQuery, example.ID)
		if err != nil {
			r.logger.Error("Failed to get example tags", zap.Error(err))
			return nil, fmt.Errorf("failed to get example tags: %w", err)
		}

		var tags []string
		for tagRows.Next() {
			var tag string
			if err := tagRows.Scan(&tag); err != nil {
				tagRows.Close()
				r.logger.Error("Failed to scan example tag", zap.Error(err))
				return nil, fmt.Errorf("failed to scan example tag: %w", err)
			}
			tags = append(tags, tag)
		}
		tagRows.Close()

		// Filter by tag if specified
		if filter.Tag != "" {
			hasTag := false
			for _, tag := range tags {
				if tag == filter.Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		example.Value = entity.ExampleValue{
			Type:  valueType,
			Count: valueCount,
			Tags:  tags,
		}

		examples = append(examples, &example)
	}

	return examples, nil
}
