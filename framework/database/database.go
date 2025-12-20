package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/axiomod/axiomod/platform/observability"

	"go.uber.org/zap"
)

// TransactionFunc is a function that executes within a transaction
type TransactionFunc func(ctx context.Context, tx *sql.Tx) error

// DB is a wrapper around sql.DB with transaction support
type DB struct {
	db     *sql.DB
	logger *observability.Logger
}

// New creates a new DB instance
func New(db *sql.DB, logger *observability.Logger) *DB {
	return &DB{
		db:     db,
		logger: logger,
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
func Connect(driverName, dataSourceName string, maxOpenConns, maxIdleConns int, connMaxLifetime time.Duration, logger *observability.Logger) (*DB, error) {
	// Open a connection to the database
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		logger.Error("Failed to open database connection", zap.Error(err))
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxIdleConns)
	db.SetConnMaxLifetime(connMaxLifetime)

	// Verify the connection
	if err := db.Ping(); err != nil {
		logger.Error("Failed to ping database", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Connected to database",
		zap.String("driver", driverName),
		zap.Int("maxOpenConns", maxOpenConns),
		zap.Int("maxIdleConns", maxIdleConns),
		zap.Duration("connMaxLifetime", connMaxLifetime),
	)

	return New(db, logger), nil
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
	return d.db.ExecContext(ctx, query, args...)
}

// Query executes a query that returns rows
func (d *DB) Query(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRow executes a query that is expected to return at most one row
func (d *DB) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// GetDB returns the underlying sql.DB instance
func (d *DB) GetDB() *sql.DB {
	return d.db
}
