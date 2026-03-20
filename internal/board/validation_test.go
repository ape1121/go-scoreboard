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
	require.Error(t, ValidateName(""))
	require.Error(t, ValidateName(strings.Repeat("a", MaxNameLength+1)))
	require.NoError(t, ValidateName("Weekly Tournament"))
	require.NoError(t, ValidateName("a"))
	require.NoError(t, ValidateName(strings.Repeat("a", MaxNameLength)))
}

func TestValidateNameTrimsWhitespace(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateName("  hello  "))
	require.Error(t, ValidateName("  "+strings.Repeat("a", MaxNameLength+1)+"  "))
}

func TestValidateNameCountsRunes(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateName(strings.Repeat("世", MaxNameLength)))
	require.Error(t, ValidateName(strings.Repeat("世", MaxNameLength+1)))
}

func TestValidateDescription(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateDescription(""))
	require.NoError(t, ValidateDescription(strings.Repeat("a", MaxDescriptionLength)))
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

func TestValidateScheduleBoundaries(t *testing.T) {
	t.Parallel()

	require.NoError(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: MinScheduleInterval}))
	require.NoError(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: MaxScheduleInterval}))
	require.Error(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: 0}))
	require.Error(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: -time.Second}))
}

func TestValidateScheduleRejectsFractionalSeconds(t *testing.T) {
	t.Parallel()

	require.Error(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: 1500 * time.Millisecond}))
	require.NoError(t, ValidateSchedule(&Schedule{Type: ScheduleTypeInterval, Interval: 2 * time.Second}))
}

func TestValidateNewBoardCombinesErrors(t *testing.T) {
	t.Parallel()

	err := ValidateNewBoard("", strings.Repeat("a", MaxDescriptionLength+1), &Schedule{Type: "bad"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "name must not be empty")
	require.Contains(t, err.Error(), "description must be at most")
	require.Contains(t, err.Error(), "schedule type must be")
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

func TestScheduleIntervalSeconds(t *testing.T) {
	t.Parallel()

	schedule := Schedule{Type: ScheduleTypeInterval, Interval: 7 * 24 * time.Hour}
	require.Equal(t, int64(604800), schedule.IntervalSeconds())
}
