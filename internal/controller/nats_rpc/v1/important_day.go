package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/internal/controller/nats_rpc/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/nats_rpc/v1/response"
	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/pkg/nats/nats_rpc/server"
	"github.com/goccy/go-json"
	"github.com/nats-io/nats.go"
)

func (r *V1) createImportantDay() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - createImportantDay - auth: %w", err)
		}

		var req request.CreateImportantDay
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - createImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - createImportantDay - validation: %w", err)
		}

		day, err := r.id.Create(context.Background(), userID, req.ToParams())
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - createImportantDay: %w", err)
		}

		return day, nil
	}
}

func (r *V1) getImportantDay() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - getImportantDay - auth: %w", err)
		}

		var req request.GetImportantDay
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - getImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - getImportantDay - validation: %w", err)
		}

		return r.id.Get(context.Background(), userID, req.ID)
	}
}

func (r *V1) listImportantDays() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - listImportantDays - auth: %w", err)
		}

		var req request.ListImportantDays
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - listImportantDays - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - listImportantDays - validation: %w", err)
		}

		var dayType *entity.ImportantDayType
		if req.Type != "" {
			t := entity.ImportantDayType(req.Type)
			dayType = &t
		}

		days, total, err := r.id.List(context.Background(), userID, dayType, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - listImportantDays: %w", err)
		}

		return response.ImportantDayList{ImportantDays: days, Total: total}, nil
	}
}

func (r *V1) upcomingImportantDays() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - upcomingImportantDays - auth: %w", err)
		}

		var req request.UpcomingImportantDays
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - upcomingImportantDays - decode: %w", err)
		}

		upcoming, total, err := r.id.Upcoming(context.Background(), userID, time.Now().UTC(), req.Days, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - upcomingImportantDays: %w", err)
		}

		return response.UpcomingImportantDayList{ImportantDays: upcoming, Total: total}, nil
	}
}

func (r *V1) updateImportantDay() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - updateImportantDay - auth: %w", err)
		}

		var req request.UpdateImportantDay
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - updateImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - updateImportantDay - validation: %w", err)
		}

		return r.id.Update(context.Background(), userID, req.ID, req.ToParams())
	}
}

func (r *V1) replaceImportantDayReminders() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - replaceImportantDayReminders - auth: %w", err)
		}

		var req request.ReplaceReminderRules
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - replaceImportantDayReminders - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - replaceImportantDayReminders - validation: %w", err)
		}

		rules, err := r.id.ReplaceReminderRules(context.Background(), userID, req.ID, req.ToParams())
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - replaceImportantDayReminders: %w", err)
		}

		return response.ReminderRuleList{Rules: rules}, nil
	}
}

func (r *V1) deleteImportantDay() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteImportantDay - auth: %w", err)
		}

		var req request.DeleteImportantDay
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteImportantDay - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteImportantDay - validation: %w", err)
		}

		if err = r.id.Delete(context.Background(), userID, req.ID); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteImportantDay: %w", err)
		}

		return response.DeleteStatus{Status: "deleted"}, nil
	}
}

func decodeNATS(data []byte, dest any) error {
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, dest)
}
