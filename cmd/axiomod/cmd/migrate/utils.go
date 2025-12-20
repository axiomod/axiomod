package migrate

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq" // PostgreSQL driver
)

// ensureDatabaseExists checks if the database exists and creates it if not.
// Note: This might require specific database privileges.
func ensureDatabaseExists(dsn string) error {
	// Parse DSN to get connection details without dbname
	// This is a simplified example for PostgreSQL
	parts := strings.Split(dsn, "?")
	connStr := parts[0]
	dbName := ""
	driverName := "postgres" // Assuming postgres, adjust if needed

	// Extract driver and dbname
	if strings.Contains(connStr, "://") {
		driverAndConn := strings.SplitN(connStr, "://", 2)
		driverName = driverAndConn[0]
		connStr = driverAndConn[1]
	}

	// Find dbname (specific to postgres DSN format)
	if driverName == "postgres" {
		// Example DSN: postgres://user:password@host:port/dbname?sslmode=disable
		// Or: host=localhost port=5432 user=postgres password=secret dbname=test sslmode=disable
		if strings.Contains(dsn, "dbname=") {
			dbNameParts := strings.Split(dsn, "dbname=")
			if len(dbNameParts) > 1 {
				dbName = strings.Split(dbNameParts[1], " ")[0]
				dbName = strings.Split(dbName, "?")[0]
			}
		}
		// Construct connection string to default 'postgres' db
		connStr = strings.Replace(dsn, "dbname="+dbName, "dbname=postgres", 1)
	} else {
		// Add logic for other database types if needed
		return fmt.Errorf("database driver '%s' not fully supported for auto-creation check", driverName)
	}

	if dbName == "" {
		return fmt.Errorf("could not parse database name from DSN")
	}

	// Connect to the default database (e.g., postgres)
	db, err := sql.Open(driverName, connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to default database: %w", err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		return fmt.Errorf("failed to ping default database: %w", err)
	}

	// Check if the target database exists
	var exists bool
	// Query varies by database type
	query := fmt.Sprintf("SELECT EXISTS (SELECT 1 FROM pg_database WHERE datname = '%s')", dbName) // PostgreSQL specific
	err = db.QueryRow(query).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check if database '%s' exists: %w", dbName, err)
	}

	// Create database if it doesn't exist
	if !exists {
		fmt.Printf("Database '%s' does not exist. Creating...\n", dbName)
		// CREATE DATABASE command varies by database type
		_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName)) // PostgreSQL specific
		if err != nil {
			return fmt.Errorf("failed to create database '%s': %w", dbName, err)
		}
		fmt.Printf("Database '%s' created successfully.\n", dbName)
	}

	return nil
}
