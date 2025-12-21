package migrate

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/axiomod/axiomod/framework/config"
	_ "github.com/lib/pq" // PostgreSQL driver
)

// getDSN loads the configuration and returns the database DSN.
func getDSN() (string, error) {
	cfg, err := config.Load("")
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	dbCfg := cfg.Database
	if dbCfg.Driver == "postgres" || dbCfg.Driver == "postgresql" {
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode), nil
	}

	return "", fmt.Errorf("unsupported database driver: %s", dbCfg.Driver)
}

// ensureDatabaseExists checks if the database exists and creates it if not.
func ensureDatabaseExists(dsn string) error {
	// Parse DSN to get base connection string and DB name
	// DSN format: postgres://user:password@host:port/dbname?sslmode=...

	// Remove protocol
	dsnWithoutProto := strings.TrimPrefix(dsn, "postgres://")

	// Split at / to get host:port and dbname?query
	parts := strings.Split(dsnWithoutProto, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid DSN format: missing /")
	}

	// Get base part (user:password@host:port)
	basePart := parts[0]

	// Get db part (dbname?sslmode=...)
	dbPart := parts[1]
	dbName := strings.Split(dbPart, "?")[0]

	// Build base DSN to connect to 'postgres' database
	baseDSN := fmt.Sprintf("postgres://%s/postgres?sslmode=disable", basePart)

	db, err := sql.Open("postgres", baseDSN)
	if err != nil {
		return fmt.Errorf("failed to open connection to base postgres: %w", err)
	}
	defer db.Close()

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", dbName)
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to query database existence: %w", err)
	}

	if !exists {
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
		if err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		fmt.Printf("Created database: %s\n", dbName)
	}

	return nil
}
