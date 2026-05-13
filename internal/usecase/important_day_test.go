package usecase_test

import (
	"context"
	"testing"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/evrone/go-clean-template/internal/usecase/importantday"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newImportantDayUseCase(t *testing.T) (*importantday.UseCase, *MockImportantDayRepo, *MockReminderRuleRepo, *MockReminderJobRepo, *MockUserSettingsRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	dayRepo := NewMockImportantDayRepo(ctrl)
	ruleRepo := NewMockReminderRuleRepo(ctrl)
	jobRepo := NewMockReminderJobRepo(ctrl)
	settingsRepo := NewMockUserSettingsRepo(ctrl)
	useCase := importantday.New(dayRepo, ruleRepo, jobRepo, settingsRepo)

	return useCase, dayRepo, ruleRepo, jobRepo, settingsRepo
}

func TestImportantDayGetReminderRules(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		expected := []entity.ReminderRule{{ID: "rule-id-123", UserID: "user-id-123", ImportantDayID: "day-id-123", OffsetDays: 7}}

		uc, dayRepo, ruleRepo, _, _ := newImportantDayUseCase(t)
		dayRepo.EXPECT().
			GetByID(context.Background(), "user-id-123", "day-id-123").
			Return(entity.ImportantDay{ID: "day-id-123", UserID: "user-id-123"}, nil)
		ruleRepo.EXPECT().
			GetForImportantDay(context.Background(), "user-id-123", "day-id-123").
			Return(expected, nil)

		rules, err := uc.GetReminderRules(context.Background(), "user-id-123", "day-id-123")

		require.NoError(t, err)
		assert.Equal(t, expected, rules)
	})

	t.Run("important day not found", func(t *testing.T) {
		t.Parallel()

		uc, dayRepo, _, _, _ := newImportantDayUseCase(t)
		dayRepo.EXPECT().
			GetByID(context.Background(), "user-id-123", "missing-day").
			Return(entity.ImportantDay{}, entity.ErrImportantDayNotFound)

		rules, err := uc.GetReminderRules(context.Background(), "user-id-123", "missing-day")

		require.ErrorIs(t, err, entity.ErrImportantDayNotFound)
		assert.Nil(t, rules)
	})

	t.Run("rule repo error", func(t *testing.T) {
		t.Parallel()

		uc, dayRepo, ruleRepo, _, _ := newImportantDayUseCase(t)
		dayRepo.EXPECT().
			GetByID(context.Background(), "user-id-123", "day-id-123").
			Return(entity.ImportantDay{ID: "day-id-123", UserID: "user-id-123"}, nil)
		ruleRepo.EXPECT().
			GetForImportantDay(context.Background(), "user-id-123", "day-id-123").
			Return(nil, errInternalServErr)

		rules, err := uc.GetReminderRules(context.Background(), "user-id-123", "day-id-123")

		require.ErrorIs(t, err, errInternalServErr)
		assert.Nil(t, rules)
	})
}
