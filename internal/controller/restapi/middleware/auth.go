package middleware

import (
	"net/http"
	"strings"

	"github.com/alfariesh/backend-memora/internal/controller/restapi/v1/response"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/gofiber/fiber/v2"
)

const _bearerParts = 2

// Auth returns a JWT authentication middleware for Fiber.
func Auth(jwtManager *jwt.Manager) func(*fiber.Ctx) error {
	return func(ctx *fiber.Ctx) error {
		header := ctx.Get("Authorization")
		if header == "" {
			return ctx.Status(http.StatusUnauthorized).JSON(response.Error{
				Error:   "missing_authorization_header",
				Message: "missing authorization header",
			})
		}

		parts := strings.SplitN(header, " ", _bearerParts)
		if len(parts) != _bearerParts || parts[0] != "Bearer" {
			return ctx.Status(http.StatusUnauthorized).JSON(response.Error{
				Error:   "invalid_authorization_header",
				Message: "invalid authorization header format",
			})
		}

		userID, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			return ctx.Status(http.StatusUnauthorized).JSON(response.Error{
				Error:   "invalid_or_expired_token",
				Message: "invalid or expired token",
			})
		}

		ctx.Locals("userID", userID)

		return ctx.Next()
	}
}
