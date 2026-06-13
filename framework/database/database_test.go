package database

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/observability"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite" // pure-Go sqlite driver for transaction tests
)

func testDeps(t *testing.T) (*config.Config, *observability.Logger, *observability.Metrics) {
	t.Helper()
	cfg := &config.Config{}
	logger, err := observability.NewLogger(cfg)
	require.NoError(t, err)
	metrics, err := observability.NewMetrics(cfg, logger)
	require.NoError(t, err)
	return cfg, logger, metrics
}

func TestDB(t *testing.T) {
	cfg, logger, metrics := testDeps(t)

	t.Run("New DB", func(t *testing.T) {
		sqlDB := &sql.DB{}
		db := New(sqlDB, logger, metrics, cfg)
		assert.NotNil(t, db)
		assert.Equal(t, sqlDB, db.GetDB())
	})
}

func TestBuildDSN(t *testing.T) {
	base := config.DatabaseConfig{
		Host:     "db.internal",
		Port:     5432,
		User:     "svc",
		Password: "s3cret",
		Name:     "appdb",
		SSLMode:  "disable",
	}

	tests := []struct {
		name   string
		mutate func(c *config.DatabaseConfig)
		want   string
	}{
		{
			"postgres keyword format",
			func(c *config.DatabaseConfig) { c.Driver = "postgres" },
			"host=db.internal port=5432 user=svc password=s3cret dbname=appdb sslmode=disable",
		},
		{
			"mysql tcp format with tls disabled",
			func(c *config.DatabaseConfig) { c.Driver = "mysql"; c.Port = 3306 },
			"svc:s3cret@tcp(db.internal:3306)/appdb?parseTime=true&tls=false",
		},
		{
			"mysql tls required",
			func(c *config.DatabaseConfig) { c.Driver = "mysql"; c.Port = 3306; c.SSLMode = "require" },
			"svc:s3cret@tcp(db.internal:3306)/appdb?parseTime=true&tls=true",
		},
		{
			"mysql unknown ssl mode omits tls param",
			func(c *config.DatabaseConfig) { c.Driver = "mysql"; c.Port = 3306; c.SSLMode = "custom" },
			"svc:s3cret@tcp(db.internal:3306)/appdb?parseTime=true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbCfg := base
			tt.mutate(&dbCfg)
			assert.Equal(t, tt.want, buildDSN(dbCfg))
		})
	}
}

func TestRedactedDSN(t *testing.T) {
	for _, driver := range []string{"postgres", "mysql"} {
		t.Run(driver, func(t *testing.T) {
			dbCfg := config.DatabaseConfig{
				Driver: driver, Host: "h", Port: 1, User: "u",
				Password: "topsecret", Name: "n", SSLMode: "disable",
			}
			redacted := redactedDSN(dbCfg)
			assert.NotContains(t, redacted, "topsecret", "password must never appear in redacted DSN")
			assert.Contains(t, redacted, "****")
		})
	}
}

func TestConnectDriverRegistration(t *testing.T) {
	cfg, logger, metrics := testDeps(t)

	t.Run("unknown driver fails at open", func(t *testing.T) {
		cfg.Database = config.DatabaseConfig{Driver: "oracle", Host: "127.0.0.1", Port: 1}
		_, err := Connect(cfg, logger, metrics, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown driver")
	})

	// Both supported drivers must be registered in this package: the error
	// has to come from the dial, never from "unknown driver".
	for _, driver := range []string{"postgres", "mysql"} {
		t.Run(driver+" driver is registered", func(t *testing.T) {
			cfg.Database = config.DatabaseConfig{
				Driver: driver, Host: "127.0.0.1", Port: 1,
				User: "u", Password: "p", Name: "d", SSLMode: "disable",
			}
			_, err := Connect(cfg, logger, metrics, nil)
			require.Error(t, err, "no database listens on port 1")
			assert.NotContains(t, err.Error(), "unknown driver")
		})
	}
}

// openSQLiteDB returns a framework DB wrapper over an in-memory SQLite
// database, letting transaction and query paths run without external infra.
func openSQLiteDB(t *testing.T) *DB {
	t.Helper()
	cfg, logger, metrics := testDeps(t)

	sqlDB, err := sql.Open("sqlite", "file:"+strings.ReplaceAll(t.Name(), "/", "_")+"?mode=memory&cache=shared")
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db := New(sqlDB, logger, metrics, cfg)
	_, err = db.Exec(context.Background(), "CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT)")
	require.NoError(t, err)
	return db
}

func TestWithTransaction(t *testing.T) {
	ctx := context.Background()

	t.Run("commit on success", func(t *testing.T) {
		db := openSQLiteDB(t)
		err := db.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			_, err := tx.ExecContext(ctx, "INSERT INTO items (name) VALUES (?)", "kept")
			return err
		})
		require.NoError(t, err)

		var count int
		require.NoError(t, db.QueryRow(ctx, "SELECT COUNT(*) FROM items").Scan(&count))
		assert.Equal(t, 1, count)
	})

	t.Run("rollback on error", func(t *testing.T) {
		db := openSQLiteDB(t)
		err := db.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
			if _, err := tx.ExecContext(ctx, "INSERT INTO items (name) VALUES (?)", "discarded"); err != nil {
				return err
			}
			return assert.AnError
		})
		assert.ErrorIs(t, err, assert.AnError)

		var count int
		require.NoError(t, db.QueryRow(ctx, "SELECT COUNT(*) FROM items").Scan(&count))
		assert.Equal(t, 0, count, "failed transaction must roll back")
	})
}

func TestQueryHelpers(t *testing.T) {
	ctx := context.Background()
	db := openSQLiteDB(t)

	_, err := db.Exec(ctx, "INSERT INTO items (name) VALUES (?), (?)", "a", "b")
	require.NoError(t, err)

	rows, err := db.Query(ctx, "SELECT name FROM items ORDER BY name")
	require.NoError(t, err)
	defer rows.Close()

	var names []string
	for rows.Next() {
		var name string
		require.NoError(t, rows.Scan(&name))
		names = append(names, name)
	}
	require.NoError(t, rows.Err())
	assert.Equal(t, []string{"a", "b"}, names)
}
