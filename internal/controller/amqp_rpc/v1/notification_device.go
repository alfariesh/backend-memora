package v1

import (
	"context"
	"fmt"

	"github.com/alfariesh/backend-memora/internal/controller/amqp_rpc/v1/request"
	"github.com/alfariesh/backend-memora/internal/controller/amqp_rpc/v1/response"
	"github.com/alfariesh/backend-memora/pkg/rabbitmq/rmq_rpc/server"
	amqp "github.com/rabbitmq/amqp091-go"
)

func (r *V1) listNotifications() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - listNotifications - auth: %w", err)
		}

		var req request.ListNotifications
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - listNotifications - decode: %w", err)
		}

		notifications, total, err := r.n.List(context.Background(), userID, req.UnreadOnly, req.Limit, req.Offset)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - listNotifications: %w", err)
		}

		return response.NotificationList{Notifications: notifications, Total: total}, nil
	}
}

func (r *V1) markNotificationRead() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - markNotificationRead - auth: %w", err)
		}

		var req request.MarkNotificationRead
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - markNotificationRead - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - markNotificationRead - validation: %w", err)
		}

		return r.n.MarkRead(context.Background(), userID, req.ID)
	}
}

func (r *V1) markAllNotificationsRead() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, _, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - markAllNotificationsRead - auth: %w", err)
		}

		if err = r.n.MarkAllRead(context.Background(), userID); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - markAllNotificationsRead: %w", err)
		}

		return response.DeleteStatus{Status: "updated"}, nil
	}
}

func (r *V1) registerDevice() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - registerDevice - auth: %w", err)
		}

		var req request.RegisterDevice
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - registerDevice - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - registerDevice - validation: %w", err)
		}

		return r.d.Register(context.Background(), userID, req.Token, req.Platform, req.Name)
	}
}

func (r *V1) deleteDevice() server.CallHandler {
	return func(d *amqp.Delivery) (any, error) {
		userID, data, err := extractUserID(d, r.j)
		if err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteDevice - auth: %w", err)
		}

		var req request.DeleteDevice
		if err = decodeAMQP(data, &req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteDevice - decode: %w", err)
		}

		if err = r.v.Struct(req); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteDevice - validation: %w", err)
		}

		if err = r.d.Delete(context.Background(), userID, req.ID); err != nil {
			return nil, fmt.Errorf("amqp_rpc - V1 - deleteDevice: %w", err)
		}

		return response.DeleteStatus{Status: "deleted"}, nil
	}
}
