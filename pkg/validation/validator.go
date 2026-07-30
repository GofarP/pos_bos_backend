package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

type ValidationError struct {
	Errors map[string]string
}

func (e *ValidationError) Error() string {
	return "validation failed"
}

func FormatValidationError(err error) error {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		valErrors := make(map[string]string)
		for _, e := range validationErrors {
			field := strings.ToLower(e.Field())
			switch e.Tag() {
			case "required":
				valErrors[field] = field + " is required"
			case "email":
				valErrors[field] = "invalid email format"
			case "min":
				valErrors[field] = field + " minimum length is " + e.Param()
			default:
				valErrors[field] = e.Error()
			}
		}
		return &ValidationError{Errors: valErrors}
	}
	return err
}
