package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/controller/amqp_rpc/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/amqp_rpc/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/rabbitmq/rmq_rpc/server"
	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *V1) createImportantDay() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - createImportantDay - auth: %w", err)
		}

		var req request.CreateImportantDay
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - createImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - createImportantDay - validation: %w", err)
		}

		day, err := r.id.Create(context.Background(), userID, req.ToParams())
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - createImportantDay: %w", err)
		}

		return day, nil
	}
}

func (r *V1) getImportantDay() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - getImportantDay - auth: %w", err)
		}

		var req request.GetImportantDay
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - getImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - getImportantDay - validation: %w", err)
		}

		return r.id.Get(context.Background(), userID, req.ID)
	}
}

func (r *V1) listImportantDays() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - listImportantDays - auth: %w", err)
		}

		var req request.ListImportantDays
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - listImportantDays - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - listImportantDays - validation: %w", err)
		}

		var dayType *entity.ImportantDayType
		if req.Type != "" {
			t := entity.ImportantDayType(req.Type)
			dayType = &t
		}

		days, total, err := r.id.List(context.Background(), userID, dayType, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - listImportantDays: %w", err)
		}

		return response.ImportantDayList{ImportantDays: days, Total: total}, nil
	}
}

func (r *V1) upcomingImportantDays() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - upcomingImportantDays - auth: %w", err)
		}

		var req request.UpcomingImportantDays
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - upcomingImportantDays - decode: %w", err)
		}

		upcoming, total, err := r.id.Upcoming(context.Background(), userID, time.Now().UTC(), req.Days, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - upcomingImportantDays: %w", err)
		}

		return response.UpcomingImportantDayList{ImportantDays: upcoming, Total: total}, nil
	}
}

func (r *V1) updateImportantDay() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - updateImportantDay - auth: %w", err)
		}

		var req request.UpdateImportantDay
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - updateImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - updateImportantDay - validation: %w", err)
		}

		return r.id.Update(context.Background(), userID, req.ID, req.ToParams())
	}
}

func (r *V1) replaceImportantDayReminders() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - replaceImportantDayReminders - auth: %w", err)
		}

		var req request.ReplaceReminderRules
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - replaceImportantDayReminders - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - replaceImportantDayReminders - validation: %w", err)
		}

		rules, err := r.id.ReplaceReminderRules(context.Background(), userID, req.ID, req.ToParams())
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - replaceImportantDayReminders: %w", err)
		}

		return response.ReminderRuleList{Rules: rules}, nil
	}
}

func (r *V1) deleteImportantDay() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteImportantDay - auth: %w", err)
		}

		var req request.DeleteImportantDay
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteImportantDay - validation: %w", err)
		}

		if err = r.id.Delete(context.Background(), userID, req.ID); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteImportantDay: %w", err)
		}

		return response.DeleteStatus{Status: "deleted"}, nil
	}
}

func decodeAMQP(data []byte, dest any) error {
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, dest)
}
