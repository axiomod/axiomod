# Database Patterns

## Connection

`framework/database/database.go`:

```go
db, err := database.Connect(cfg, logger, metrics, health)
```

Handles: connection pool settings, ping, health check registration.

## Transactions

Use `WithTransaction` for automatic rollback/commit:

```go
err := db.WithTransaction(ctx, func(tx *sql.Tx) error {
    if _, err := tx.ExecContext(ctx, query1, args...); err != nil {
        return err  // Auto rollback
    }
    if _, err := tx.ExecContext(ctx, query2, args...); err != nil {
        return err  // Auto rollback
    }
    return nil  // Auto commit
})
```

## Query Wrappers

`Exec`, `Query`, `QueryRow` wrappers automatically:
- Record duration metrics via `DBQueryDuration`
- Detect and log slow queries (threshold: `Database.SlowQueryThreshold`, default 200ms)

## Migrations

Via CLI:

```
axiomod migrate up       # Apply pending migrations
axiomod migrate down     # Rollback last migration
axiomod migrate create   # Create new migration file
axiomod migrate force    # Force set version
axiomod migrate version  # Show current version
```

Uses `golang-migrate/v4`.

## Rules

1. Always use `WithTransaction` for multi-statement operations
2. Use parameterized queries -- never string concatenation (SQL injection prevention)
3. Configure `SlowQueryThreshold` per environment
4. Register database health check via `health.RegisterCheck`
5. Close connections gracefully on shutdown
