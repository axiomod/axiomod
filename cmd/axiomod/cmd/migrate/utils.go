package migrate

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/axiomod/axiomod/framework/config"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	_ "github.com/lib/pq"              // PostgreSQL driver
)

// getDSN loads the configuration and returns the golang-migrate database URL
// for the configured driver.
func getDSN() (string, error) {
	cfg, err := config.Load("")
	if err != nil {
		return "", fmt.Errorf("failed to load config: %w", err)
	}

	dbCfg := cfg.Database
	switch dbCfg.Driver {
	case "postgres", "postgresql":
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name, dbCfg.SSLMode), nil
	case "mysql":
		// golang-migrate's mysql driver expects mysql://<go-sql-driver DSN>.
		return fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s",
			dbCfg.User, dbCfg.Password, dbCfg.Host, dbCfg.Port, dbCfg.Name), nil
	default:
		return "", fmt.Errorf("unsupported database driver: %s (supported: postgres, mysql)", dbCfg.Driver)
	}
}

// ensureDatabaseExists checks if the target database exists and creates it if
// not, dispatching on the DSN scheme.
func ensureDatabaseExists(dsn string) error {
	switch {
	case strings.HasPrefix(dsn, "postgres://"):
		return ensurePostgresDatabase(dsn)
	case strings.HasPrefix(dsn, "mysql://"):
		return ensureMySQLDatabase(dsn)
	default:
		return fmt.Errorf("unsupported DSN scheme in %q", redactDSN(dsn))
	}
}

// ensurePostgresDatabase connects to the maintenance database and creates the
// target database when missing.
func ensurePostgresDatabase(dsn string) error {
	// DSN format: postgres://user:password@host:port/dbname?sslmode=...
	dsnWithoutProto := strings.TrimPrefix(dsn, "postgres://")

	parts := strings.Split(dsnWithoutProto, "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid DSN format: missing /")
	}

	basePart := parts[0] // user:password@host:port
	dbName := strings.Split(parts[1], "?")[0]

	baseDSN := fmt.Sprintf("postgres://%s/postgres?sslmode=disable", basePart)

	db, err := sql.Open("postgres", baseDSN)
	if err != nil {
		return fmt.Errorf("failed to open connection to base postgres: %w", err)
	}
	defer db.Close()

	var exists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", dbName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to query database existence: %w", err)
	}

	if !exists {
		if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %q", dbName)); err != nil {
			return fmt.Errorf("failed to create database %s: %w", dbName, err)
		}
		fmt.Printf("Created database: %s\n", dbName)
	}

	return nil
}

// ensureMySQLDatabase connects without a schema selected and creates the
// target database when missing.
func ensureMySQLDatabase(dsn string) error {
	// DSN format: mysql://user:password@tcp(host:port)/dbname
	native := strings.TrimPrefix(dsn, "mysql://")

	idx := strings.LastIndex(native, "/")
	if idx < 0 {
		return fmt.Errorf("invalid DSN format: missing /")
	}

	basePart := native[:idx] // user:password@tcp(host:port)
	dbName := strings.Split(native[idx+1:], "?")[0]

	db, err := sql.Open("mysql", basePart+"/")
	if err != nil {
		return fmt.Errorf("failed to open base mysql connection: %w", err)
	}
	defer db.Close()

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s`", dbName)); err != nil {
		return fmt.Errorf("failed to create database %s: %w", dbName, err)
	}

	return nil
}

// redactDSN masks the credentials section of a database URL for safe output.
func redactDSN(dsn string) string {
	schemeEnd := strings.Index(dsn, "://")
	at := strings.LastIndex(dsn, "@")
	if schemeEnd < 0 || at < 0 || at < schemeEnd {
		return dsn
	}
	return dsn[:schemeEnd+3] + "****" + dsn[at:]
}
