package entity_test

import (
	"testing"
	"time"

	"github.com/evrone/go-clean-template/internal/entity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportantDay_NextOccurrence(t *testing.T) {
	t.Parallel()

	day := entity.ImportantDay{
		EventMonth:   5,
		EventDay:     13,
		Timezone:     entity.DefaultTimezone,
		ReminderTime: entity.DefaultReminderTime,
	}

	occurrence, err := day.NextOccurrence(time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	assert.Equal(t, "2026-05-13", occurrence.Format("2006-01-02"))
}

func TestImportantDay_NextOccurrenceRollsToNextYear(t *testing.T) {
	t.Parallel()

	day := entity.ImportantDay{
		EventMonth:   5,
		EventDay:     13,
		Timezone:     entity.DefaultTimezone,
		ReminderTime: entity.DefaultReminderTime,
	}

	occurrence, err := day.NextOccurrence(time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	assert.Equal(t, "2027-05-13", occurrence.Format("2006-01-02"))
}

func TestImportantDay_LeapDayFallsBackToFeb28(t *testing.T) {
	t.Parallel()

	day := entity.ImportantDay{
		EventMonth:   2,
		EventDay:     29,
		Timezone:     entity.DefaultTimezone,
		ReminderTime: entity.DefaultReminderTime,
	}

	occurrence, err := day.NextOccurrence(time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	assert.Equal(t, "2026-02-28", occurrence.Format("2006-01-02"))
}

func TestImportantDay_AnniversaryFor(t *testing.T) {
	t.Parallel()

	year := 2020
	day := entity.ImportantDay{EventYear: &year}

	anniversary := day.AnniversaryFor(time.Date(2026, 5, 13, 0, 0, 0, 0, time.UTC))

	require.NotNil(t, anniversary)
	assert.Equal(t, 6, *anniversary)
}

func TestNormalizeReminderRules_Defaults(t *testing.T) {
	t.Parallel()

	rules := entity.NormalizeReminderRules(nil)

	require.Len(t, rules, 3)
	assert.Equal(t, []int{7, 1, 0}, []int{rules[0].OffsetDays, rules[1].OffsetDays, rules[2].OffsetDays})
	assert.Equal(t, entity.DefaultReminderChannels, rules[0].Channels)
}

func TestNormalizeReminderRulesWithChannels(t *testing.T) {
	t.Parallel()

	rules := entity.NormalizeReminderRulesWithChannels(nil, []entity.ReminderChannel{entity.ReminderChannelInApp})

	assert.Len(t, rules, 3)
	assert.Equal(t, []entity.ReminderChannel{entity.ReminderChannelInApp}, rules[0].Channels)
}

func TestNormalizeUserSettings(t *testing.T) {
	t.Parallel()

	params := entity.UserSettingsParams{}
	require.NoError(t, entity.NormalizeUserSettings(&params))

	assert.Equal(t, entity.DefaultTimezone, params.Timezone)
	assert.Equal(t, entity.DefaultReminderTime, params.ReminderTime)
	assert.Equal(t, entity.DefaultReminderChannels, params.NotificationChannels)
}

func TestNormalizeUserSettings_EmptyChannels(t *testing.T) {
	t.Parallel()

	params := entity.UserSettingsParams{NotificationChannels: []entity.ReminderChannel{}}
	require.NoError(t, entity.NormalizeUserSettings(&params))

	assert.Empty(t, params.NotificationChannels)
}

func TestNormalizeUserSettings_InvalidTimezone(t *testing.T) {
	t.Parallel()

	params := entity.UserSettingsParams{Timezone: "Invalid/Zone"}

	assert.ErrorIs(t, entity.NormalizeUserSettings(&params), entity.ErrInvalidUserSettings)
}

func TestFilterReminderChannels(t *testing.T) {
	t.Parallel()

	filtered := entity.FilterReminderChannels(
		[]entity.ReminderChannel{entity.ReminderChannelEmail, entity.ReminderChannelInApp, entity.ReminderChannelPush},
		[]entity.ReminderChannel{entity.ReminderChannelInApp, entity.ReminderChannelPush},
	)

	assert.Equal(t, []entity.ReminderChannel{entity.ReminderChannelInApp, entity.ReminderChannelPush}, filtered)
}

func TestIsExpoPushToken(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		token string
		want  bool
	}{
		{name: "expo token", token: "ExpoPushToken[abc123]", want: true},
		{name: "legacy exponent token", token: "ExponentPushToken[abc123]", want: true},
		{name: "empty payload", token: "ExpoPushToken[]", want: false},
		{name: "missing suffix", token: "ExpoPushToken[abc123", want: false},
		{name: "native token", token: "fcm-token", want: false},
		{name: "empty", token: "", want: false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.want, entity.IsExpoPushToken(tc.token))
		})
	}
}
