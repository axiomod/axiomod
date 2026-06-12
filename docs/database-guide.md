# Database Guide

The Axiomod framework provides a robust way to interact with SQL databases,
specifically MySQL and PostgreSQL, with **Ent as the default ORM** and plain
`database/sql` available behind a configuration switch.

## 1. Configuration

Database settings are managed in `configs/service_default.yaml`.

```yaml
database:
  driver: "postgres" # or mysql
  host: "localhost"
  port: 5432
  user: "axiomod"
  password: "password"
  name: "axiomod"
  sslMode: "disable"
  orm: "ent" # Options: ent (default), sql (plain database/sql repositories)
  maxOpenConns: 25
  maxIdleConns: 5
  connMaxLifetime: 15 # minutes
  slowQueryThreshold: 200 # milliseconds
```

### Environment Overrides

The standard `APP_` prefix applies (Viper key dots become underscores):

```bash
export APP_DATABASE_DRIVER=postgres
export APP_DATABASE_HOST=db.internal
export APP_DATABASE_USER=axiomod
export APP_DATABASE_PASSWORD=change-me
export APP_DATABASE_NAME=axiomod
export APP_DATABASE_ORM=ent
```

## 2. The ORM Layer (Ent — default)

With `database.orm: "ent"`, repositories are implemented with the
[Ent](https://entgo.io) ORM. The framework pieces:

- **Schemas** live in `platform/ent/schema/`; regenerate the client with
  `go generate ./platform/ent/...` after changes.
- **`ent.NewClientFromDB(driver, db)`** (`platform/ent/factory.go`) wraps
  the framework's managed `*sql.DB` pool, so pooling, slow-query logging,
  and metrics keep flowing through one place.
- The reference implementation is
  `examples/example/infrastructure/persistence/example_ent_repository.go`,
  selected by `ProvideRepository` in `examples/example/module.go` when a
  database plugin is enabled. `repo.Migrate(ctx)` creates the schema at
  startup.

Set `database.orm: "sql"` to use the plain `database/sql` implementation
(`example_sql_repository.go`) instead — the repository interface is
identical, so use cases never notice the difference. With **no** database
plugin enabled, domain modules fall back to thread-safe in-memory
repositories so the server boots with zero infrastructure.

## 3. The database/sql Wrapper

The framework provides a `database.DB` wrapper in `framework/database`
(`database.Connect` registers both supported drivers, builds driver-specific
DSNs, configures the pool, pings, and registers a health check).

```go
type MyRepository struct {
    db *database.DB
}

func NewMyRepository(db *database.DB) *MyRepository {
    return &MyRepository{db: db}
}
```

### Executing Queries

The wrapper's `Exec`, `Query`, and `QueryRow` record duration metrics and log
slow queries (threshold: `database.slowQueryThreshold`).

```go
func (r *MyRepository) GetByID(ctx context.Context, id string) (*User, error) {
    var user User
    query := "SELECT id, name FROM users WHERE id = $1"
    err := r.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.Name)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

## 4. Transaction Management

The framework simplifies transaction management with the `WithTransaction`
helper:

```go
func (r *MyRepository) UpdateUser(ctx context.Context, user *User) error {
    return r.db.WithTransaction(ctx, func(ctx context.Context, tx *sql.Tx) error {
        _, err := tx.ExecContext(ctx, "UPDATE users SET name = $1 WHERE id = $2", user.Name, user.ID)
        if err != nil {
            return err
        }

        _, err = tx.ExecContext(ctx, "INSERT INTO audit_logs (...) VALUES (...)")
        return err
    })
}
```

If the function returns an error, the transaction is automatically rolled
back. Otherwise, it is committed.

## 5. Plugins (MySQL/PostgreSQL)

The `postgres` and `mysql` plugins establish the managed connection at
startup via `database.Connect` and register the `database` health check.
Enable exactly one in your configuration:

```yaml
plugins:
  enabled:
    postgres: true
    mysql: false
```

Note: `plugins.enabled` is a map of plugin name to boolean. The connection
parameters come from the `database:` section above.

## 6. Database Migrations

Automate your schema changes with the built-in migration tool using
`golang-migrate`. Both **postgres** and **mysql** are supported (selected by
`database.driver`); `migrate up` creates the target database when missing.

```bash
# Create a new migration pair (up/down) under ./migrations
axiomod migrate create add_users_table

# Apply all pending migrations
axiomod migrate up

# Roll back the last migration step (or N steps)
axiomod migrate down [N]

# Force a specific version (useful for dirty states)
axiomod migrate force 20231010120000

# Show current version and dirty flag
axiomod migrate version
```

### Best Practices

- Always test migrations (`up` and `down`) in development.
- Use descriptive names for your migration files.
- Backup your database before running migrations in production.
- Ent's `Schema.Create` is convenient for development; prefer explicit
  golang-migrate migrations for production schema changes.
