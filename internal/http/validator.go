package http

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validate is a singleton validator instance
var Validate = validator.New()

// init initializes the validator with custom settings
func init() {
	// Register validation for ensuring status is a valid enum value
	Validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})
}

// parseValidationErrors converts validator errors into a map for error responses
func parseValidationErrors(err error) map[string]interface{} {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		errors := make(map[string]interface{})
		
		for _, e := range validationErrors {
			// Use the json field name as the key
			field := e.Field()
			
			var message string
			switch e.Tag() {
			case "required":
				message = "This field is required"
			case "min":
				message = "Value is below minimum allowed value"
			case "max":
				message = "Value exceeds maximum allowed value"
			case "oneof":
				message = "Value must be one of the allowed values: " + e.Param()
			default:
				message = "Invalid value"
			}
			
			errors[field] = message
		}
		
		return errors
	}
	
	// If it's not a validation error or can't be parsed, return generic message
	return map[string]interface{}{
		"message": "Invalid input",
	}
}
