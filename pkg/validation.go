package pkg

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

func ParseValidationErrors(err error) string {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return err.Error()
	}

	msgs := make([]string, len(ve))
	for i, fe := range ve {
		msgs[i] = fieldErrorMessage(fe)
	}
	return strings.Join(msgs, "; ")
}

func fieldErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("field '%s' is required", fe.Field())
	case "email":
		return fmt.Sprintf("field '%s' must be a valid email", fe.Field())
	case "gt":
		return fmt.Sprintf("field '%s' must be greater than %s", fe.Field(), fe.Param())
	default:
		return fmt.Sprintf("field '%s' is invalid", fe.Field())
	}
}
