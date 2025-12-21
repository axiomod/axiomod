package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"

	"go.uber.org/zap"
)

// TransactionFunc is a function that executes within a transaction
type TransactionFunc func(ctx context.Context, tx *sql.Tx) error

// DB is a wrapper around sql.DB with transaction support
type DB struct {
	db      *sql.DB
	logger  *observability.Logger
	metrics *observability.Metrics
	cfg     *config.Config
}

// New creates a new DB instance
func New(db *sql.DB, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config) *DB {
	return &DB{
		db:      db,
		logger:  logger,
		metrics: metrics,
		cfg:     cfg,
	}
}

// WithTransaction executes the given function within a transaction
func (d *DB) WithTransaction(ctx context.Context, fn TransactionFunc) error {
	// Start a transaction
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		d.logger.Error("Failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Execute the function
	if err := fn(ctx, tx); err != nil {
		// Rollback the transaction on error
		if rbErr := tx.Rollback(); rbErr != nil {
			d.logger.Error("Failed to rollback transaction", zap.Error(rbErr))
			return fmt.Errorf("failed to rollback transaction: %w (original error: %v)", rbErr, err)
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		d.logger.Error("Failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Connect establishes a connection to the database
func Connect(cfg *config.Config, logger *observability.Logger, metrics *observability.Metrics, health *health.Health) (*DB, error) {
	dbCfg := cfg.Database
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.Password, dbCfg.Name, dbCfg.SSLMode)

	// Open a connection to the database
	db, err := sql.Open(dbCfg.Driver, dsn)
	if err != nil {
		logger.Error("Failed to open database connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set default pool settings if not provided
	if dbCfg.MaxOpenConns == 0 {
		dbCfg.MaxOpenConns = 25
	}
	if dbCfg.MaxIdleConns == 0 {
		dbCfg.MaxIdleConns = 25
	}
	if dbCfg.ConnMaxLifetime == 0 {
		dbCfg.ConnMaxLifetime = 5 // 5 minutes
	}

	// Set connection pool settings
	db.SetMaxOpenConns(dbCfg.MaxOpenConns)
	db.SetMaxIdleConns(dbCfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(dbCfg.ConnMaxLifetime) * time.Minute)

	// Verify the connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to database",
		zap.String("driver", dbCfg.Driver),
		zap.Int("maxOpenConns", dbCfg.MaxOpenConns),
		zap.Int("maxIdleConns", dbCfg.MaxIdleConns),
		zap.Int("connMaxLifetimeMin", dbCfg.ConnMaxLifetime),
	)

	// Register health check
	if health != nil {
		health.RegisterCheck("database", func() error {
			return db.Ping()
		})
	}

	return New(db, logger, metrics, cfg), nil
}

// Close closes the database connection
func (d *DB) Close() error {
	if err := d.db.Close(); err != nil {
		d.logger.Error("Failed to close database connection", zap.Error(err))
		return fmt.Errorf("failed to close database connection: %w", err)
	}
	d.logger.Info("Closed database connection")
	return nil
}

// Exec executes a query without returning any rows
func (d *DB) Exec(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	start := time.Now()
	res, err := d.db.ExecContext(ctx, query, args...)
	d.recordQuery(query, "exec", start, err)
	return res, err
}

// Query executes a query that returns rows
func (d *DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.db.QueryContext(ctx, query, args...)
	d.recordQuery(query, "query", start, err)
	return rows, err
}

// QueryRow executes a query that is expected to return at most one row
func (d *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	start := time.Now()
	row := d.db.QueryRowContext(ctx, query, args...)
	// Note: We can't easily check for error until Scan is called,
	// but we record the duration anyway.
	d.recordQuery(query, "query_row", start, nil)
	return row
}

func (d *DB) recordQuery(query, queryType string, start time.Time, err error) {
	duration := time.Since(start)

	// Record metrics
	status := "success"
	if err != nil {
		status = "error"
	}
	if d.metrics != nil && d.metrics.DBQueryDuration != nil {
		d.metrics.DBQueryDuration.WithLabelValues(queryType, status).Observe(duration.Seconds())
	}

	// Log slow queries
	threshold := 200 * time.Millisecond // Default 200ms
	if d.cfg != nil && d.cfg.Database.SlowQueryThreshold > 0 {
		threshold = time.Duration(d.cfg.Database.SlowQueryThreshold) * time.Millisecond
	}

	if duration > threshold {
		d.logger.Warn("Slow database query detected",
			zap.String("query", query),
			zap.String("type", queryType),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
	}
}

// GetDB returns the underlying sql.DB instance
func (d *DB) GetDB() *sql.DB {
	return d.db
}
