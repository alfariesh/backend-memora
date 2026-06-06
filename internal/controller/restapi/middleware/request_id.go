package middleware

import (
	"crypto/rand"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	// RequestIDHeader is the primary request correlation header.
	RequestIDHeader = "X-Request-ID"
	// CorrelationIDHeader is accepted as an alias for upstream compatibility.
	CorrelationIDHeader = "X-Correlation-ID"
	// RequestIDLocalKey stores the sanitized request ID in Fiber locals.
	RequestIDLocalKey = "requestID"

	maxRequestIDLength = 128
)

// RequestID propagates a safe caller-provided request ID or creates a new one.
func RequestID() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		requestID := firstValidRequestID(ctx.Get(RequestIDHeader), ctx.Get(CorrelationIDHeader))
		if requestID == "" {
			requestID = newRequestID()
		}

		ctx.Locals(RequestIDLocalKey, requestID)
		ctx.Set(RequestIDHeader, requestID)
		ctx.Set(CorrelationIDHeader, requestID)

		return ctx.Next()
	}
}

// RequestIDFromContext returns the sanitized request ID stored by RequestID.
func RequestIDFromContext(ctx *fiber.Ctx) string {
	requestID, _ := ctx.Locals(RequestIDLocalKey).(string)

	return requestID
}

func firstValidRequestID(values ...string) string {
	for _, value := range values {
		if isValidRequestID(value) {
			return value
		}
	}

	return ""
}

func isValidRequestID(value string) bool {
	if value == "" || len(value) > maxRequestIDLength {
		return false
	}

	for _, char := range value {
		switch {
		case char >= 'a' && char <= 'z':
		case char >= 'A' && char <= 'Z':
		case char >= '0' && char <= '9':
		case char == '-', char == '_', char == '.', char == ':':
		default:
			return false
		}
	}

	return true
}

func newRequestID() string {
	var bytes [16]byte
	if _, err := rand.Read(bytes[:]); err != nil {
		return fmt.Sprintf("req-%x", time.Now().UnixNano())
	}

	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}
