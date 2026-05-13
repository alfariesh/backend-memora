package usecase_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase/device"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newDeviceUseCase(t *testing.T) (*device.UseCase, *MockDeviceTokenRepo, *MockPushSender) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repo := NewMockDeviceTokenRepo(ctrl)
	pushSender := NewMockPushSender(ctrl)
	useCase := device.New(repo, device.PushSender(pushSender))

	return useCase, repo, pushSender
}

func TestDeviceTestPush(t *testing.T) {
	t.Parallel()

	deviceToken := entity.DeviceToken{
		ID:     "device-id-123",
		UserID: "user-id-123",
		Token:  "ExpoPushToken[test]",
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		uc, repo, pushSender := newDeviceUseCase(t)
		repo.EXPECT().ListActiveByUser(context.Background(), "user-id-123").Return([]entity.DeviceToken{deviceToken}, nil)
		pushSender.EXPECT().
			Send(context.Background(), deviceToken.Token, "Custom title", "Custom body", gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _, _ string, data map[string]string) (string, error) {
				assert.Equal(t, "test_push", data["type"])
				assert.Equal(t, deviceToken.ID, data["device_id"])
				assert.NotEmpty(t, data["sent_at"])

				return "ticket-id-123", nil
			})

		result, err := uc.TestPush(context.Background(), "user-id-123", "device-id-123", "Custom title", "Custom body")

		require.NoError(t, err)
		assert.Equal(t, "device-id-123", result.DeviceID)
		assert.Equal(t, "ticket-id-123", result.TicketID)
		assert.NotZero(t, result.SentAt)
	})

	t.Run("device not found", func(t *testing.T) {
		t.Parallel()

		uc, repo, _ := newDeviceUseCase(t)
		repo.EXPECT().ListActiveByUser(context.Background(), "user-id-123").Return([]entity.DeviceToken{}, nil)

		_, err := uc.TestPush(context.Background(), "user-id-123", "missing-device", "", "")

		require.ErrorIs(t, err, entity.ErrDeviceTokenNotFound)
	})

	t.Run("push device not registered deactivates token", func(t *testing.T) {
		t.Parallel()

		uc, repo, pushSender := newDeviceUseCase(t)
		repo.EXPECT().ListActiveByUser(context.Background(), "user-id-123").Return([]entity.DeviceToken{deviceToken}, nil)
		pushSender.EXPECT().
			Send(context.Background(), deviceToken.Token, "Memora test", "Push notifications are working.", gomock.Any()).
			Return("", fmt.Errorf("%w: inactive", entity.ErrPushDeviceNotRegistered))
		repo.EXPECT().Deactivate(context.Background(), "user-id-123", "device-id-123", gomock.Any()).Return(nil)

		_, err := uc.TestPush(context.Background(), "user-id-123", "device-id-123", "", "")

		require.ErrorIs(t, err, entity.ErrPushDeviceNotRegistered)
	})

	t.Run("push send failure", func(t *testing.T) {
		t.Parallel()

		uc, repo, pushSender := newDeviceUseCase(t)
		repo.EXPECT().ListActiveByUser(context.Background(), "user-id-123").Return([]entity.DeviceToken{deviceToken}, nil)
		pushSender.EXPECT().
			Send(context.Background(), deviceToken.Token, "Memora test", "Push notifications are working.", gomock.Any()).
			Return("", errInternalServErr)

		_, err := uc.TestPush(context.Background(), "user-id-123", "device-id-123", "", "")

		require.ErrorIs(t, err, entity.ErrPushSendFailed)
	})

	t.Run("push sender not configured", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		repo := NewMockDeviceTokenRepo(ctrl)
		uc := device.New(repo)

		_, err := uc.TestPush(context.Background(), "user-id-123", "device-id-123", "", "")

		require.ErrorIs(t, err, entity.ErrPushSenderNotConfigured)
	})
}
