package persistence

import (
	"context"
	"database/sql"
	"testing"

	"github.com/axiomod/axiomod/examples/example/entity"
	"github.com/axiomod/axiomod/examples/example/repository"
	"github.com/axiomod/axiomod/framework/observability"
	"github.com/axiomod/axiomod/platform/ent"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	_ "modernc.org/sqlite" // pure-Go sqlite driver for tests
)

// newEntTestRepository spins up an Ent repository backed by an in-memory
// SQLite database, exercising the same generated code paths as Postgres/MySQL.
func newEntTestRepository(t *testing.T) *ExampleEntRepository {
	t.Helper()

	db, err := sql.Open("sqlite", "file:enttest?mode=memory&cache=shared&_pragma=foreign_keys(1)")
	require.NoError(t, err)
	// SQLite in-memory databases vanish when the last connection closes.
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	client, err := ent.NewClientFromDB("sqlite", db)
	require.NoError(t, err)

	logger := &observability.Logger{Logger: zap.NewNop()}
	repo := NewExampleEntRepository(client, logger)
	require.NoError(t, repo.Migrate(context.Background()))
	return repo
}

func newTestExample(name, valueType string, tags []string) *entity.Example {
	value := entity.NewExampleValue(valueType, 1, tags)
	return entity.NewExample(name, "description for "+name, value)
}

func TestExampleEntRepositoryCRUD(t *testing.T) {
	repo := newEntTestRepository(t)
	ctx := context.Background()

	created := newTestExample("first", "premium", []string{"alpha"})
	require.NoError(t, repo.Create(ctx, created))

	t.Run("duplicate id", func(t *testing.T) {
		assert.ErrorIs(t, repo.Create(ctx, created), repository.ErrDuplicateID)
	})

	t.Run("get by id", func(t *testing.T) {
		got, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, created.Name, got.Name)
		assert.Equal(t, created.Value.Type, got.Value.Type)
		assert.Equal(t, created.Value.Tags, got.Value.Tags)
	})

	t.Run("get missing", func(t *testing.T) {
		_, err := repo.GetByID(ctx, "missing")
		assert.ErrorIs(t, err, repository.ErrExampleNotFound)
	})

	t.Run("update", func(t *testing.T) {
		created.Update("renamed", "new description", entity.NewExampleValue("basic", 2, []string{"beta"}))
		require.NoError(t, repo.Update(ctx, created))

		got, err := repo.GetByID(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, "renamed", got.Name)
		assert.Equal(t, "basic", got.Value.Type)
	})

	t.Run("update missing", func(t *testing.T) {
		ghost := newTestExample("ghost", "premium", nil)
		assert.ErrorIs(t, repo.Update(ctx, ghost), repository.ErrExampleNotFound)
	})

	t.Run("delete", func(t *testing.T) {
		require.NoError(t, repo.Delete(ctx, created.ID))
		_, err := repo.GetByID(ctx, created.ID)
		assert.ErrorIs(t, err, repository.ErrExampleNotFound)
		assert.ErrorIs(t, repo.Delete(ctx, created.ID), repository.ErrExampleNotFound)
	})
}

func TestExampleEntRepositoryList(t *testing.T) {
	repo := newEntTestRepository(t)
	ctx := context.Background()

	seed := []*entity.Example{
		newTestExample("a", "premium", []string{"x"}),
		newTestExample("b", "premium", []string{"y"}),
		newTestExample("c", "basic", []string{"x", "y"}),
	}
	for _, e := range seed {
		require.NoError(t, repo.Create(ctx, e))
	}

	tests := []struct {
		name   string
		filter repository.ExampleFilter
		want   int
	}{
		{"all", repository.ExampleFilter{}, 3},
		{"by value type", repository.ExampleFilter{ValueType: "premium"}, 2},
		{"by name", repository.ExampleFilter{Name: "c"}, 1},
		{"by tag", repository.ExampleFilter{Tag: "x"}, 2},
		{"tag and type", repository.ExampleFilter{ValueType: "basic", Tag: "x"}, 1},
		{"paginated", repository.ExampleFilter{Limit: 2}, 2},
		{"offset past end", repository.ExampleFilter{Limit: 2, Offset: 5}, 0},
		{"no match", repository.ExampleFilter{Name: "zzz"}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.List(ctx, tt.filter)
			require.NoError(t, err)
			assert.Len(t, got, tt.want)
		})
	}
}
