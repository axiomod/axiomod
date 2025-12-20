package entity

import (
	"time"

	"github.com/google/uuid"
)

// Example represents an example entity in the domain
type Example struct {
	ID          string
	Name        string
	Description string
	Value       ExampleValue
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewExample creates a new Example entity
func NewExample(name, description string, value ExampleValue) *Example {
	now := time.Now()
	return &Example{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Value:       value,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

// Update updates the Example entity
func (e *Example) Update(name, description string, value ExampleValue) {
	e.Name = name
	e.Description = description
	e.Value = value
	e.UpdatedAt = time.Now()
}

// Validate validates the Example entity
func (e *Example) Validate() error {
	if e.Name == "" {
		return ErrEmptyName
	}
	if e.Description == "" {
		return ErrEmptyDescription
	}
	return nil
}

// Errors for Example entity validation
var (
	ErrEmptyName        = NewDomainError("example.empty_name", "Name cannot be empty")
	ErrEmptyDescription = NewDomainError("example.empty_description", "Description cannot be empty")
)

// DomainError represents a domain error
type DomainError struct {
	Code    string
	Message string
}

// Error returns the error message
func (e DomainError) Error() string {
	return e.Message
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) DomainError {
	return DomainError{
		Code:    code,
		Message: message,
	}
}
