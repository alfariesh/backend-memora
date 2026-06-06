package middleware

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const permissionsPolicy = "geolocation=(), microphone=(), camera=()"

// SecurityHeadersConfig contains browser security header settings.
type SecurityHeadersConfig struct {
	Enabled               bool
	HSTSEnabled           bool
	HSTSMaxAge            int
	HSTSIncludeSubdomains bool
	HSTSPreload           bool
}

// SecurityHeaders returns a Fiber middleware that sets browser hardening headers.
func SecurityHeaders(cfg SecurityHeadersConfig) fiber.Handler {
	if !cfg.Enabled {
		return nextOnly
	}

	return func(ctx *fiber.Ctx) error {
		ctx.Set(fiber.HeaderXContentTypeOptions, "nosniff")
		ctx.Set(fiber.HeaderXFrameOptions, "DENY")
		ctx.Set(fiber.HeaderReferrerPolicy, "no-referrer")
		ctx.Set(fiber.HeaderPermissionsPolicy, permissionsPolicy)
		ctx.Set("Origin-Agent-Cluster", "?1")
		ctx.Set("X-DNS-Prefetch-Control", "off")
		ctx.Set("X-Download-Options", "noopen")
		ctx.Set("X-Permitted-Cross-Domain-Policies", "none")

		if cfg.HSTSEnabled && cfg.HSTSMaxAge > 0 {
			ctx.Set(fiber.HeaderStrictTransportSecurity, buildHSTSHeader(cfg))
		}

		return ctx.Next()
	}
}

func buildHSTSHeader(cfg SecurityHeadersConfig) string {
	var header strings.Builder
	fmt.Fprintf(&header, "max-age=%d", cfg.HSTSMaxAge)

	if cfg.HSTSIncludeSubdomains {
		header.WriteString("; includeSubDomains")
	}

	if cfg.HSTSPreload {
		header.WriteString("; preload")
	}

	return header.String()
}
