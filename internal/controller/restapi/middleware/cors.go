package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// CORSConfig contains browser cross-origin policy settings.
type CORSConfig struct {
	Enabled          bool
	AllowedOrigins   string
	AllowedMethods   string
	AllowedHeaders   string
	ExposedHeaders   string
	AllowCredentials bool
	MaxAge           int
}

// CORS returns a Fiber CORS middleware with exact-origin defaults.
func CORS(cfg CORSConfig) fiber.Handler {
	if !cfg.Enabled {
		return nextOnly
	}

	allowedOrigins := normalizeCommaSeparated(cfg.AllowedOrigins)
	corsConfig := cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     normalizeCommaSeparated(cfg.AllowedMethods),
		AllowHeaders:     normalizeCommaSeparated(cfg.AllowedHeaders),
		AllowCredentials: cfg.AllowCredentials,
		ExposeHeaders:    normalizeCommaSeparated(cfg.ExposedHeaders),
		MaxAge:           cfg.MaxAge,
	}

	if allowedOrigins == "" {
		corsConfig.AllowOriginsFunc = func(string) bool { return false }
	}

	return cors.New(corsConfig)
}

func normalizeCommaSeparated(value string) string {
	parts := strings.Split(value, ",")
	normalized := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			normalized = append(normalized, part)
		}
	}

	return strings.Join(normalized, ",")
}

func nextOnly(ctx *fiber.Ctx) error {
	return ctx.Next()
}
