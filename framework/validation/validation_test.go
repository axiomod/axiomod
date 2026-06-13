package validation

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestValidateMessagesAndJSONNames(t *testing.T) {
	v := New()

	type form struct {
		Email string `json:"email" validate:"required,email"`
		Name  string `json:"name" validate:"required,min=3,max=10"`
		Kind  string `json:"kind" validate:"required,oneof=basic premium"`
		Site  string `json:"site" validate:"omitempty,url"`
		Age   int    `json:"age" validate:"min=18"`
	}

	errsByField := func(in form) map[string]ValidationError {
		fieldErrs, err := v.Validate(in)
		require.ErrorIs(t, err, ErrValidationFailed)
		out := map[string]ValidationError{}
		for _, fe := range fieldErrs {
			out[fe.Field] = fe
		}
		return out
	}

	t.Run("messages per tag", func(t *testing.T) {
		errs := errsByField(form{Email: "nope", Name: "ab", Kind: "other", Site: "not-a-url", Age: 12})

		assert.Equal(t, "Invalid email format", errs["email"].Message)
		assert.Equal(t, "Must be at least 3 characters long", errs["name"].Message)
		assert.Equal(t, "Must be one of: basic premium", errs["kind"].Message)
		assert.Equal(t, "Invalid URL format", errs["site"].Message)
		assert.Equal(t, "Must be at least 18", errs["age"].Message)
	})

	t.Run("json tag names are reported", func(t *testing.T) {
		errs := errsByField(form{})
		_, hasJSONName := errs["email"]
		assert.True(t, hasJSONName, "field names must come from json tags, got %v", errs)
	})

	t.Run("max length message", func(t *testing.T) {
		errs := errsByField(form{Email: "a@b.co", Name: "waytoolongname", Kind: "basic", Age: 20})
		assert.Equal(t, "Must be at most 10 characters long", errs["name"].Message)
	})

	t.Run("valid input has no errors", func(t *testing.T) {
		fieldErrs, err := v.Validate(form{Email: "a@b.co", Name: "abc", Kind: "premium", Age: 20})
		assert.NoError(t, err)
		assert.Empty(t, fieldErrs)
	})
}

func TestRegisterCustomValidation(t *testing.T) {
	v := New()
	require.NoError(t, v.RegisterCustomValidation("evenlen", func(fl validator.FieldLevel) bool {
		return len(fl.Field().String())%2 == 0
	}))

	type payload struct {
		Code string `json:"code" validate:"evenlen"`
	}

	fieldErrs, err := v.Validate(payload{Code: "abc"})
	require.ErrorIs(t, err, ErrValidationFailed)
	require.Len(t, fieldErrs, 1)
	assert.Equal(t, "evenlen", fieldErrs[0].Tag)
	assert.Equal(t, "Invalid value", fieldErrs[0].Message)

	_, err = v.Validate(payload{Code: "abcd"})
	assert.NoError(t, err)
}
