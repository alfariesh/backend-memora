package request

import "github.com/alfariesh/backend-memora/internal/entity"

// Register -.
type Register struct {
	Username string `json:"username" validate:"required,min=3,max=255" example:"johndoe"`
	Email    string `json:"email"    validate:"required,email"         example:"john@example.com"`
	Password string `json:"password" validate:"required,min=8,max=72"  example:"secret123"`
} // @name v1.Register

// Login -.
type Login struct {
	Email    string `json:"email"    validate:"required,email" example:"john@example.com"`
	Password string `json:"password" validate:"required,max=72" example:"secret123"`
} // @name v1.Login

// RefreshToken -.
type RefreshToken struct {
	RefreshToken string `json:"refresh_token" validate:"required" example:"C-Kt0pA3..."`
} // @name v1.RefreshToken

// ChangePassword -.
type ChangePassword struct {
	CurrentPassword string `json:"current_password" validate:"required,max=72"    example:"secret123"`
	NewPassword     string `json:"new_password"     validate:"required,min=8,max=72" example:"newsecret123"`
} // @name v1.ChangePassword

// UpdateUserSettings -.
type UpdateUserSettings struct {
	Timezone             string                   `json:"timezone"              validate:"omitempty,max=64"                  example:"Asia/Jakarta"`
	ReminderTime         string                   `json:"reminder_time"         validate:"omitempty,datetime=15:04"          example:"09:00"`
	NotificationChannels []entity.ReminderChannel `json:"notification_channels" validate:"omitempty,dive,oneof=email in_app push" example:"email"`
} // @name v1.UpdateUserSettings

func (r UpdateUserSettings) ToParams() entity.UserSettingsParams {
	return entity.UserSettingsParams{
		Timezone:             r.Timezone,
		ReminderTime:         r.ReminderTime,
		NotificationChannels: r.NotificationChannels,
	}
}
