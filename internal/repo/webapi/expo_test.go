package webapi

import (
	"errors"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/goccy/go-json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseExpoTicket(t *testing.T) {
	t.Parallel()

	ticket, err := parseExpoTicket(json.RawMessage(`{"status":"ok","id":"ticket-id"}`))

	require.NoError(t, err)
	assert.Equal(t, "ticket-id", ticket.ID)
	assert.Equal(t, "ok", ticket.Status)
}

func TestParseExpoTicketList(t *testing.T) {
	t.Parallel()

	ticket, err := parseExpoTicket(json.RawMessage(`[{"status":"ok","id":"ticket-id"}]`))

	require.NoError(t, err)
	assert.Equal(t, "ticket-id", ticket.ID)
}

func TestExpoDeviceNotRegisteredError(t *testing.T) {
	t.Parallel()

	ticket, err := parseExpoTicket(json.RawMessage(`{"status":"error","message":"Device not registered","details":{"error":"DeviceNotRegistered"}}`))

	require.NoError(t, err)
	assert.Equal(t, "DeviceNotRegistered", ticket.Details["error"])
	assert.True(t, errors.Is(expoTicketError(ticket), entity.ErrPushDeviceNotRegistered))
}
