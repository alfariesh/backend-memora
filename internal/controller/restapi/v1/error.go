package v1

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/response"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

func errorResponse(ctx *fiber.Ctx, code int, msg string) error {
	return ctx.Status(code).JSON(response.Error{
		Error:   errorCode(msg),
		Message: msg,
	})
}

func validationErrorResponse(ctx *fiber.Ctx, err error) error {
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return errorResponse(ctx, http.StatusBadRequest, "invalid request body")
	}

	fields := make(map[string]string, len(validationErrors))
	for _, fieldErr := range validationErrors {
		fields[validationField(fieldErr)] = validationMessage(fieldErr)
	}

	return ctx.Status(http.StatusBadRequest).JSON(response.Error{
		Error:   "validation_error",
		Message: "validation failed",
		Fields:  fields,
	})
}

func validationField(fieldErr validator.FieldError) string {
	field := fieldErr.Namespace()
	if field == "" {
		field = fieldErr.Field()
	}

	if idx := strings.Index(field, "."); idx >= 0 {
		field = field[idx+1:]
	}

	if field == "" {
		return fieldErr.StructField()
	}

	return field
}

func validationMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email"
	case "min":
		if isStringKind(fieldErr.Kind()) {
			return "must be at least " + fieldErr.Param() + " characters"
		}

		return "must be at least " + fieldErr.Param()
	case "max":
		if isStringKind(fieldErr.Kind()) {
			return "must be at most " + fieldErr.Param() + " characters"
		}

		return "must be at most " + fieldErr.Param()
	case "oneof":
		return "must be one of: " + strings.ReplaceAll(fieldErr.Param(), " ", ", ")
	case "datetime":
		return "must match format " + fieldErr.Param()
	default:
		return "is invalid"
	}
}

func isStringKind(kind reflect.Kind) bool {
	return kind == reflect.String
}

func errorCode(message string) string {
	var builder strings.Builder
	lastUnderscore := false

	for _, char := range message {
		switch {
		case char >= 'a' && char <= 'z':
			builder.WriteRune(char)
			lastUnderscore = false
		case char >= 'A' && char <= 'Z':
			builder.WriteRune(char + ('a' - 'A'))
			lastUnderscore = false
		case char >= '0' && char <= '9':
			builder.WriteRune(char)
			lastUnderscore = false
		default:
			if builder.Len() > 0 && !lastUnderscore {
				builder.WriteByte('_')
				lastUnderscore = true
			}
		}
	}

	return strings.Trim(builder.String(), "_")
}
