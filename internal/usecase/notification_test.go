package usecase_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newNotificationUseCase(t *testing.T) (*notification.UseCase, *MockNotificationRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	repo := NewMockNotificationRepo(ctrl)
	useCase := notification.New(repo)

	return useCase, repo
}

func TestNotificationListCapsLimitAndNormalizesOffset(t *testing.T) {
	t.Parallel()

	uc, notificationRepo := newNotificationUseCase(t)
	notificationRepo.EXPECT().
		List(context.Background(), "user-id-123", repo.NotificationFilter{
			UnreadOnly: true,
			Limit:      100,
			Offset:     0,
		}).
		Return([]entity.Notification{}, 0, nil)

	notifications, total, err := uc.List(context.Background(), "user-id-123", true, 1000, -1)

	require.NoError(t, err)
	assert.Empty(t, notifications)
	assert.Zero(t, total)
}

func TestNotificationCountUnread(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		uc, repo := newNotificationUseCase(t)
		repo.EXPECT().CountUnread(context.Background(), "user-id-123").Return(3, nil)

		count, err := uc.CountUnread(context.Background(), "user-id-123")

		require.NoError(t, err)
		assert.Equal(t, 3, count)
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		uc, repo := newNotificationUseCase(t)
		repo.EXPECT().CountUnread(context.Background(), "user-id-123").Return(0, errInternalServErr)

		count, err := uc.CountUnread(context.Background(), "user-id-123")

		require.ErrorIs(t, err, errInternalServErr)
		assert.Zero(t, count)
	})
}
