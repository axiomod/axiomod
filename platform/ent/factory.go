package ent

import (
	"database/sql"
	"fmt"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
)

// NewClientFromDB wraps an existing database/sql connection — typically the
// framework's managed pool from framework/database — in an Ent client, so
// pooling, slow-query logging, and metrics keep flowing through one place.
func NewClientFromDB(driverName string, db *sql.DB) (*Client, error) {
	var d string
	switch driverName {
	case "postgres", "postgresql":
		d = dialect.Postgres
	case "mysql":
		d = dialect.MySQL
	case "sqlite", "sqlite3":
		d = dialect.SQLite
	default:
		return nil, fmt.Errorf("unsupported driver for ent client: %s", driverName)
	}
	drv := entsql.OpenDB(d, db)
	return NewClient(Driver(drv)), nil
}
