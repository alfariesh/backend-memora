package entity

import "errors"

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserAlreadyExists       = errors.New("user already exists")
	ErrInvalidCredentials      = errors.New("invalid credentials")
	ErrTaskNotFound            = errors.New("task not found")
	ErrTaskForbidden           = errors.New("task does not belong to user")
	ErrInvalidTransition       = errors.New("invalid status transition")
	ErrImportantDayNotFound    = errors.New("important day not found")
	ErrImportantDayForbidden   = errors.New("important day does not belong to user")
	ErrInvalidImportantDayDate = errors.New("invalid important day date")
	ErrReminderRuleNotFound    = errors.New("reminder rule not found")
	ErrReminderJobNotFound     = errors.New("reminder job not found")
	ErrNotificationNotFound    = errors.New("notification not found")
	ErrDeviceTokenNotFound     = errors.New("device token not found")
	ErrInvalidDeviceToken      = errors.New("invalid device token")
	ErrPushDeviceNotRegistered = errors.New("push device not registered")
)
