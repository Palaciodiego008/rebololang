package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate is the global validator instance
var validate *validator.Validate

func init() {
	validate = validator.New()
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Tag     string `json:"tag"`
	Value   string `json:"value"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// ValidateStruct validates a struct and returns user-friendly errors
func ValidateStruct(v interface{}) error {
	err := validate.Struct(v)
	if err == nil {
		return nil
	}

	// Convert validator errors to our custom format
	var validationErrors ValidationErrors
	
	if errs, ok := err.(validator.ValidationErrors); ok {
		for _, e := range errs {
			validationErrors = append(validationErrors, ValidationError{
				Field:   e.Field(),
				Tag:     e.Tag(),
				Value:   fmt.Sprintf("%v", e.Value()),
				Message: getErrorMessage(e),
			})
		}
	}

	return validationErrors
}

// getErrorMessage returns a user-friendly error message
func getErrorMessage(e validator.FieldError) string {
	field := e.Field()
	
	switch e.Tag() {
	case "required":
		return fmt.Sprintf("%s es requerido", field)
	case "email":
		return fmt.Sprintf("%s debe ser un email válido", field)
	case "min":
		return fmt.Sprintf("%s debe tener al menos %s caracteres", field, e.Param())
	case "max":
		return fmt.Sprintf("%s debe tener máximo %s caracteres", field, e.Param())
	case "len":
		return fmt.Sprintf("%s debe tener exactamente %s caracteres", field, e.Param())
	case "gt":
		return fmt.Sprintf("%s debe ser mayor que %s", field, e.Param())
	case "gte":
		return fmt.Sprintf("%s debe ser mayor o igual a %s", field, e.Param())
	case "lt":
		return fmt.Sprintf("%s debe ser menor que %s", field, e.Param())
	case "lte":
		return fmt.Sprintf("%s debe ser menor o igual a %s", field, e.Param())
	case "alpha":
		return fmt.Sprintf("%s solo puede contener letras", field)
	case "alphanum":
		return fmt.Sprintf("%s solo puede contener letras y números", field)
	case "numeric":
		return fmt.Sprintf("%s debe ser numérico", field)
	case "url":
		return fmt.Sprintf("%s debe ser una URL válida", field)
	case "uri":
		return fmt.Sprintf("%s debe ser una URI válida", field)
	case "eqfield":
		return fmt.Sprintf("%s debe ser igual a %s", field, e.Param())
	case "nefield":
		return fmt.Sprintf("%s no debe ser igual a %s", field, e.Param())
	default:
		return fmt.Sprintf("%s no es válido", field)
	}
}

// ValidationErrorsToMap converts validation errors to a map for easy template rendering
func ValidationErrorsToMap(err error) map[string]string {
	result := make(map[string]string)
	
	if validationErrors, ok := err.(ValidationErrors); ok {
		for _, e := range validationErrors {
			result[e.Field] = e.Message
		}
	}
	
	return result
}

