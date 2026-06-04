package v1

import (
	"github.com/alfariesh/backend-memora/internal/usecase"
	"github.com/alfariesh/backend-memora/pkg/jwt"
	"github.com/alfariesh/backend-memora/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	u  usecase.User
	tk usecase.Task
	id usecase.ImportantDay
	n  usecase.Notification
	d  usecase.DeviceToken
	j  *jwt.Manager
	l  logger.Interface
	v  *validator.Validate
}
