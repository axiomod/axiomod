# Database Guide

The Axiomod framework provides a robust way to interact with SQL databases, specifically MySQL and PostgreSQL.

## 1. Configuration

Database settings are managed in `service_default.yaml`.

```yaml
database:
  driver: mysql  # or postgres
  host: localhost
  port: 3306
  user: root
  password: your-password
  name: axiomod
  sslMode: disable
  maxOpenConns: 25
  maxIdleConns: 5
  connMaxLifetime: 300  # seconds
```

### Environment Overrides

- `DB_DRIVER`
- `DB_HOST`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`

## 2. Usage

The framework provides a `database.DB` wrapper in `internal/framework/database`.

### Getting a Connection

The `database.Module` provides a `*database.DB` instance via dependency injection.

```go
type MyRepository struct {
    db *database.DB
}

func NewMyRepository(db *database.DB) *MyRepository {
    return &MyRepository{db: db}
}
```

### Executing Queries

The wrapper exposes standard `sql.DB` methods but with built-in logging and error handling.

```go
func (r *MyRepository) GetByID(ctx context.Context, id string) (*User, error) {
    var user User
    query := "SELECT id, name FROM users WHERE id = ?"
    err := r.db.GetDB().QueryRowContext(ctx, query, id).Scan(&user.ID, &user.Name)
    if err != nil {
        return nil, err
    }
    return &user, nil
}
```

## 3. Transaction Management

The framework simplifies transaction management with the `WithTransaction` helper.

```go
func (r *MyRepository) UpdateUser(ctx context.Context, user *User) error {
    return r.db.WithTransaction(ctx, func(tx *sql.Tx) error {
        // Perform multiple operations within the transaction
        _, err := tx.ExecContext(ctx, "UPDATE users SET name = ? WHERE id = ?", user.Name, user.ID)
        if err != nil {
            return err
        }
        
        _, err = tx.ExecContext(ctx, "INSERT INTO audit_logs (...) VALUES (...)")
        return err
    })
}
```

If the function returns an error, the transaction is automatically rolled back. Otherwise, it is committed.

## 4. Plugins (MySQL/PostgreSQL)

While the `database` package provides the wrapper, specific drivers are managed as plugins in `internal/plugins`. These plugins handle the actual connection established at startup using the `database.Connect` function.

### Enabling Database Plugins

Ensure the desired plugin is enabled in your configuration:

```yaml
plugins:
  enabled:
    - mysql  # or postgres
```
