package example

import (
	"context"
	"testing"

	"axiomod/internal/examples/example/entity"
	"axiomod/internal/examples/example/infrastructure/persistence"
	"axiomod/internal/examples/example/repository"
	"axiomod/internal/examples/example/usecase"

	"github.com/stretchr/testify/assert"
)

func TestExampleEntity(t *testing.T) {
	// Create a value object
	value := entity.NewExampleValue("test", 42, []string{"tag1", "tag2"})

	// Create an entity
	example := entity.NewExample("Test Example", "This is a test example", value)

	// Validate entity
	assert.NoError(t, example.Validate())
	assert.NotEmpty(t, example.ID)
	assert.Equal(t, "Test Example", example.Name)
	assert.Equal(t, "This is a test example", example.Description)
	assert.Equal(t, "test", example.Value.Type)
	assert.Equal(t, 42, example.Value.Count)
	assert.Contains(t, example.Value.Tags, "tag1")
	assert.Contains(t, example.Value.Tags, "tag2")

	// Test validation errors
	invalidExample := entity.NewExample("", "", value)
	err := invalidExample.Validate()
	assert.Error(t, err)
	assert.Equal(t, entity.ErrEmptyName, err)

	invalidExample.Name = "Test"
	err = invalidExample.Validate()
	assert.Error(t, err)
	assert.Equal(t, entity.ErrEmptyDescription, err)
}

func TestExampleValueObject(t *testing.T) {
	// Create a value object
	value := entity.NewExampleValue("test", 42, []string{"tag1", "tag2"})

	// Test equality
	value2 := entity.NewExampleValue("test", 42, []string{"tag1", "tag2"})
	assert.True(t, value.Equals(value2))

	// Test inequality
	value3 := entity.NewExampleValue("test", 43, []string{"tag1", "tag2"})
	assert.False(t, value.Equals(value3))

	// Test has tag
	assert.True(t, value.HasTag("tag1"))
	assert.False(t, value.HasTag("tag3"))

	// Test add tag
	value.AddTag("tag3")
	assert.True(t, value.HasTag("tag3"))

	// Test remove tag
	value.RemoveTag("tag1")
	assert.False(t, value.HasTag("tag1"))
}

func TestExampleRepository(t *testing.T) {
	// Create a repository
	repo := persistence.NewExampleMemoryRepository()

	// Create a value object
	value := entity.NewExampleValue("test", 42, []string{"tag1", "tag2"})

	// Create an entity
	example := entity.NewExample("Test Example", "This is a test example", value)

	// Test create
	ctx := context.Background()
	err := repo.Create(ctx, example)
	assert.NoError(t, err)

	// Test get by ID
	retrieved, err := repo.GetByID(ctx, example.ID)
	assert.NoError(t, err)
	assert.Equal(t, example.ID, retrieved.ID)
	assert.Equal(t, example.Name, retrieved.Name)
	assert.Equal(t, example.Description, retrieved.Description)
	assert.Equal(t, example.Value.Type, retrieved.Value.Type)
	assert.Equal(t, example.Value.Count, retrieved.Value.Count)
	assert.ElementsMatch(t, example.Value.Tags, retrieved.Value.Tags)

	// Test update
	example.Name = "Updated Example"
	example.Description = "This is an updated example"
	example.Value.Count = 43
	err = repo.Update(ctx, example)
	assert.NoError(t, err)

	// Test get after update
	retrieved, err = repo.GetByID(ctx, example.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Example", retrieved.Name)
	assert.Equal(t, "This is an updated example", retrieved.Description)
	assert.Equal(t, 43, retrieved.Value.Count)

	// Test list with filter
	filter := repository.ExampleFilter{
		ValueType: "test",
	}
	examples, err := repo.List(ctx, filter)
	assert.NoError(t, err)
	assert.Len(t, examples, 1)

	// Test delete
	err = repo.Delete(ctx, example.ID)
	assert.NoError(t, err)

	// Test get after delete
	_, err = repo.GetByID(ctx, example.ID)
	assert.Error(t, err)
	assert.Equal(t, repository.ErrExampleNotFound, err)
}

func TestCreateExampleUseCase(t *testing.T) {
	// Create a repository
	repo := persistence.NewExampleMemoryRepository()

	// Create a use case
	uc := usecase.NewCreateExampleUseCase(repo)

	// Create input
	input := usecase.CreateExampleInput{
		Name:        "Test Example",
		Description: "This is a test example",
		ValueType:   "test",
		Count:       42,
		Tags:        []string{"tag1", "tag2"},
	}

	// Execute use case
	ctx := context.Background()
	output, err := uc.Execute(ctx, input)
	assert.NoError(t, err)
	assert.NotEmpty(t, output.ID)

	// Verify entity was created
	example, err := repo.GetByID(ctx, output.ID)
	assert.NoError(t, err)
	assert.Equal(t, input.Name, example.Name)
	assert.Equal(t, input.Description, example.Description)
	assert.Equal(t, input.ValueType, example.Value.Type)
	assert.Equal(t, input.Count, example.Value.Count)
	assert.ElementsMatch(t, input.Tags, example.Value.Tags)
}

func TestGetExampleUseCase(t *testing.T) {
	// Create a repository
	repo := persistence.NewExampleMemoryRepository()

	// Create a value object
	value := entity.NewExampleValue("test", 42, []string{"tag1", "tag2"})

	// Create an entity
	example := entity.NewExample("Test Example", "This is a test example", value)

	// Add to repository
	ctx := context.Background()
	err := repo.Create(ctx, example)
	assert.NoError(t, err)

	// Create a use case
	uc := usecase.NewGetExampleUseCase(repo)

	// Create input
	input := usecase.GetExampleInput{
		ID: example.ID,
	}

	// Execute use case
	output, err := uc.Execute(ctx, input)
	assert.NoError(t, err)
	assert.Equal(t, example.ID, output.ID)
	assert.Equal(t, example.Name, output.Name)
	assert.Equal(t, example.Description, output.Description)
	assert.Equal(t, example.Value.Type, output.ValueType)
	assert.Equal(t, example.Value.Count, output.Count)
	assert.ElementsMatch(t, example.Value.Tags, output.Tags)
}
