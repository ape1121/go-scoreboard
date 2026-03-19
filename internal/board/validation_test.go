package board

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestValidateName(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateName("   "))
	require.Error(t, ValidateName(strings.Repeat("a", MaxNameLength+1)))
	require.NoError(t, ValidateName("Weekly Tournament"))
}

func TestValidateDescription(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateDescription(""))
	require.Error(t, ValidateDescription(strings.Repeat("a", MaxDescriptionLength+1)))
}

func TestValidateSchedule(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateSchedule(nil))
	require.Error(t, ValidateSchedule(&Schedule{Type: "cron", Interval: time.Minute}))
	require.Error(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: 500 * time.Millisecond}))
	require.Error(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: MaxScheduleInterval + time.Second}))
	require.NoError(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: 7 * 24 * time.Hour}))
}

func TestBoardHasSchedule(t *testing.T) {
	t.Parallel()

	require.False(t, Board{}.HasSchedule())
	require.True(t, Board{Schedule: &Schedule{Type: ScheduleTypeInterval, Interval: time.Hour}}.HasSchedule())
}

func TestBoardPeriodIsActive(t *testing.T) {
	t.Parallel()

	require.True(t, BoardPeriod{}.IsActive())

	now := time.Now().UTC()
	require.False(t, BoardPeriod{EndedAt: &now}.IsActive())
}
