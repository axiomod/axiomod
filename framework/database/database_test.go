package database

import (
	"database/sql"
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"

	"github.com/stretchr/testify/assert"
)

// We can't easily test Connect without a real DB or sqlmock.
// But we can test the DB wrapper logic using a nil or dummy sql.DB if methods don't panic.
// Actually, sql.Open doesn't connect immediately, so we can test Connect with a fake driver name.

func TestDB(t *testing.T) {
	obsCfg := &config.Config{}
	logger, _ := observability.NewLogger(obsCfg)

	t.Run("New DB", func(t *testing.T) {
		sqlDB := &sql.DB{}
		db := New(sqlDB, logger)
		assert.NotNil(t, db)
		assert.Equal(t, sqlDB, db.GetDB())
	})

	// For WithTransaction and other methods that call the underlying sql.DB,
	// we would ideally use sqlmock. Since it's not explicitly in go.mod as a dependency
	// we might want to avoid adding it if not necessary, but for database tests it's standard.
}
