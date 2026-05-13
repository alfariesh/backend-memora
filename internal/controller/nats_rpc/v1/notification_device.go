package v1

import (
	"context"
	"fmt"

	"github.com/evrone/go-clean-template/internal/controller/nats_rpc/v1/request"
	"github.com/evrone/go-clean-template/internal/controller/nats_rpc/v1/response"
	"github.com/evrone/go-clean-template/pkg/nats/nats_rpc/server"
	"github.com/nats-io/nats.go"
)

func (r *V1) listNotifications() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - listNotifications - auth: %w", err)
		}

		var req request.ListNotifications
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - listNotifications - decode: %w", err)
		}

		notifications, total, err := r.n.List(context.Background(), userID, req.UnreadOnly, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - listNotifications: %w", err)
		}

		return response.NotificationList{Notifications: notifications, Total: total}, nil
	}
}

func (r *V1) markNotificationRead() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - markNotificationRead - auth: %w", err)
		}

		var req request.MarkNotificationRead
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - markNotificationRead - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - markNotificationRead - validation: %w", err)
		}

		return r.n.MarkRead(context.Background(), userID, req.ID)
	}
}

func (r *V1) markAllNotificationsRead() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, _, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - markAllNotificationsRead - auth: %w", err)
		}

		if err = r.n.MarkAllRead(context.Background(), userID); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - markAllNotificationsRead: %w", err)
		}

		return response.DeleteStatus{Status: "updated"}, nil
	}
}

func (r *V1) registerDevice() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - registerDevice - auth: %w", err)
		}

		var req request.RegisterDevice
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - registerDevice - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - registerDevice - validation: %w", err)
		}

		return r.d.Register(context.Background(), userID, req.Token, req.Platform, req.Name)
	}
}

func (r *V1) deleteDevice() server.CallHandler {
	return func(msg *nats.Msg) (any, error) {
		userID, data, err := extractUserID(msg, r.j)
		if err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteDevice - auth: %w", err)
		}

		var req request.DeleteDevice
		if err = decodeNATS(data, &req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteDevice - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteDevice - validation: %w", err)
		}

		if err = r.d.Delete(context.Background(), userID, req.ID); err != nil {
			return nil, fmt.Errorf("nats_rpc - V1 - deleteDevice: %w", err)
		}

		return response.DeleteStatus{Status: "deleted"}, nil
	}
}
