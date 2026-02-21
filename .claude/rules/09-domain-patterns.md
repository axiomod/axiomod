# Domain Patterns (Clean Architecture)

## Entity

```go
type Example struct {
    ID          string
    Name        string
    Description string
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// Constructor generates UUID and sets timestamps
func NewExample(name, description string) (*Example, error) {
    e := &Example{
        ID:          uuid.New().String(),
        Name:        name,
        Description: description,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }
    if err := e.Validate(); err != nil {
        return nil, err
    }
    return e, nil
}

// Update method modifies fields and updates timestamp
func (e *Example) Update(name, description string) error {
    e.Name = name
    e.Description = description
    e.UpdatedAt = time.Now()
    return e.Validate()
}

// Validate returns domain errors
func (e *Example) Validate() error {
    if e.Name == "" {
        return ErrEmptyName
    }
    return nil
}
```

Domain errors in entity package:

```go
var (
    ErrEmptyName = NewDomainError("example.empty_name", "Name cannot be empty")
)
```

## Repository Interface

```go
type ExampleRepository interface {
    Create(ctx context.Context, example *entity.Example) error
    GetByID(ctx context.Context, id string) (*entity.Example, error)
    Update(ctx context.Context, example *entity.Example) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, filter ExampleFilter) ([]*entity.Example, error)
}

type ExampleFilter struct {
    Name   string
    Limit  int
    Offset int
}
```

Repository errors:

```go
var (
    ErrExampleNotFound = entity.NewDomainError("repository.example_not_found", "Example not found")
    ErrDuplicateID     = entity.NewDomainError("repository.duplicate_id", "Duplicate ID")
)
```

## Use Case (one per file)

```go
type CreateExampleInput struct {
    Name        string `json:"name" validate:"required"`
    Description string `json:"description" validate:"required"`
}

type CreateExampleOutput struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
}

type CreateExampleUseCase struct {
    repo repository.ExampleRepository
}

func NewCreateExampleUseCase(repo repository.ExampleRepository) *CreateExampleUseCase {
    return &CreateExampleUseCase{repo: repo}
}

func (uc *CreateExampleUseCase) Execute(ctx context.Context,
    input CreateExampleInput) (*CreateExampleOutput, error) {
    // Business logic here
    example, err := entity.NewExample(input.Name, input.Description)
    if err != nil {
        return nil, err
    }
    if err := uc.repo.Create(ctx, example); err != nil {
        return nil, err
    }
    return &CreateExampleOutput{
        ID: example.ID, Name: example.Name,
        Description: example.Description, CreatedAt: example.CreatedAt,
    }, nil
}
```

## Service (cross-entity logic)

```go
type ExampleDomainService struct {
    repo repository.ExampleRepository
}

func NewExampleDomainService(repo repository.ExampleRepository) *ExampleDomainService {
    return &ExampleDomainService{repo: repo}
}
```

## Infrastructure: Memory Repository

```go
type ExampleMemoryRepository struct {
    mu       sync.RWMutex
    examples map[string]*entity.Example
}

func (r *ExampleMemoryRepository) GetByID(ctx context.Context, id string) (*entity.Example, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    example, ok := r.examples[id]
    if !ok {
        return nil, repository.ErrExampleNotFound
    }
    return cloneExample(example), nil  // Deep clone to prevent shared state
}
```

## Infrastructure: Cache

```go
type ExampleCache interface {
    Get(ctx context.Context, id string) (*entity.Example, error)
    Set(ctx context.Context, example *entity.Example, ttl time.Duration) error
    Delete(ctx context.Context, id string) error
}
```

Key pattern: `"example:<id>"`

## Infrastructure: Event Publisher

```go
const (
    EventExampleCreated = "example.created"
    EventExampleUpdated = "example.updated"
    EventExampleDeleted = "example.deleted"
)
```

Topic naming: `"examples.<eventType>"`

## Validation

Use `validator/v10` struct tags on use case inputs:

```go
Name  string `json:"name" validate:"required"`
Email string `json:"email" validate:"required,email"`
Age   int    `json:"age" validate:"min=18"`
```

## Rules

1. Entities encapsulate business rules and validation
2. Repositories define interfaces only -- implementations live in `infrastructure/`
3. One use case per file with clear `Input`/`Output` structs
4. Use cases orchestrate entities and repositories
5. Services handle cross-entity business logic
6. Memory repositories must be thread-safe (`sync.RWMutex`) and deep-clone entities
