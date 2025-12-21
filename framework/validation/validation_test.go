package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=18"`
}

func TestValidator(t *testing.T) {
	v := New()

	t.Run("Valid struct", func(t *testing.T) {
		s := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   25,
		}
		errors, err := v.Validate(s)
		assert.NoError(t, err)
		assert.Nil(t, errors)
	})

	t.Run("Invalid email", func(t *testing.T) {
		s := TestStruct{
			Name:  "John Doe",
			Email: "invalid-email",
			Age:   25,
		}
		errors, err := v.Validate(s)
		assert.Error(t, err)
		assert.Equal(t, ErrValidationFailed, err)
		assert.Len(t, errors, 1)
		assert.Equal(t, "email", errors[0].Field)
		assert.Equal(t, "email", errors[0].Tag)
		assert.Equal(t, "Invalid email format", errors[0].Message)
	})

	t.Run("Missing required field", func(t *testing.T) {
		s := TestStruct{
			Email: "john@example.com",
			Age:   25,
		}
		errors, err := v.Validate(s)
		assert.Error(t, err)
		assert.Len(t, errors, 1)
		assert.Equal(t, "name", errors[0].Field)
		assert.Equal(t, "required", errors[0].Tag)
	})

	t.Run("Min age validation", func(t *testing.T) {
		s := TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   17,
		}
		errors, err := v.Validate(s)
		assert.Error(t, err)
		assert.Len(t, errors, 1)
		assert.Equal(t, "age", errors[0].Field)
		assert.Equal(t, "min", errors[0].Tag)
		assert.Equal(t, "18", errors[0].Value)
	})
}
