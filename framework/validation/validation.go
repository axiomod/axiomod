package validation

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Common errors
var (
	ErrValidationFailed = errors.New("validation failed")
)

// Validator provides validation functionality
type Validator struct {
	validator *validator.Validate
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// New creates a new Validator
func New() *Validator {
	v := validator.New()

	// Register function to get json tag names instead of struct field names
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &Validator{
		validator: v,
	}
}

// Validate validates the given struct
func (v *Validator) Validate(s interface{}) ([]ValidationError, error) {
	if err := v.validator.Struct(s); err != nil {
		var validationErrors []ValidationError

		// Convert validator errors to our custom format
		for _, err := range err.(validator.ValidationErrors) {
			validationErrors = append(validationErrors, ValidationError{
				Field:   err.Field(),
				Tag:     err.Tag(),
				Value:   err.Param(),
				Message: getErrorMessage(err),
			})
		}

		return validationErrors, ErrValidationFailed
	}

	return nil, nil
}

// RegisterCustomValidation registers a custom validation function
func (v *Validator) RegisterCustomValidation(tag string, fn validator.Func) error {
	return v.validator.RegisterValidation(tag, fn)
}

// getErrorMessage returns a human-readable error message for a validation error
func getErrorMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Invalid email format"
	case "min":
		if err.Type().Kind() == reflect.String {
			return "Must be at least " + err.Param() + " characters long"
		}
		return "Must be at least " + err.Param()
	case "max":
		if err.Type().Kind() == reflect.String {
			return "Must be at most " + err.Param() + " characters long"
		}
		return "Must be at most " + err.Param()
	case "oneof":
		return "Must be one of: " + err.Param()
	case "url":
		return "Invalid URL format"
	default:
		return "Invalid value"
	}
}
