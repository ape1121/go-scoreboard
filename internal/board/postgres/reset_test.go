package postgres

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAdvancePeriodReturnsNotDueBeforeBoundary(t *testing.T) {
	t.Parallel()

	nextStartedAt, increments, due := advancePeriod(
		time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		time.Hour,
		time.Date(2026, 3, 19, 12, 59, 59, 0, time.UTC),
	)

	require.False(t, due)
	require.True(t, nextStartedAt.IsZero())
	require.Zero(t, increments)
}

func TestAdvancePeriodReturnsSingleIntervalReset(t *testing.T) {
	t.Parallel()

	nextStartedAt, increments, due := advancePeriod(
		time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		time.Hour,
		time.Date(2026, 3, 19, 13, 5, 0, 0, time.UTC),
	)

	require.True(t, due)
	require.Equal(t, time.Date(2026, 3, 19, 13, 0, 0, 0, time.UTC), nextStartedAt)
	require.EqualValues(t, 1, increments)
}

func TestAdvancePeriodReturnsMultipleMissedIntervals(t *testing.T) {
	t.Parallel()

	nextStartedAt, increments, due := advancePeriod(
		time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		time.Hour,
		time.Date(2026, 3, 19, 15, 30, 0, 0, time.UTC),
	)

	require.True(t, due)
	require.Equal(t, time.Date(2026, 3, 19, 15, 0, 0, 0, time.UTC), nextStartedAt)
	require.EqualValues(t, 3, increments)
}

func TestAdvancePeriodExactBoundaryIsDue(t *testing.T) {
	t.Parallel()

	nextStartedAt, increments, due := advancePeriod(
		time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		time.Hour,
		time.Date(2026, 3, 19, 13, 0, 0, 0, time.UTC),
	)

	require.True(t, due)
	require.Equal(t, time.Date(2026, 3, 19, 13, 0, 0, 0, time.UTC), nextStartedAt)
	require.EqualValues(t, 1, increments)
}

func TestAdvancePeriodZeroIntervalIsNotDue(t *testing.T) {
	t.Parallel()

	_, _, due := advancePeriod(
		time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		0,
		time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC),
	)

	require.False(t, due)
}

func TestAdvancePeriodNegativeIntervalIsNotDue(t *testing.T) {
	t.Parallel()

	_, _, due := advancePeriod(
		time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC),
		-time.Hour,
		time.Date(2026, 3, 19, 14, 0, 0, 0, time.UTC),
	)

	require.False(t, due)
}

func TestAdvancePeriodLargeNumberOfMissedIntervals(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, 3, 19, 0, 0, 0, 0, time.UTC)

	nextStartedAt, increments, due := advancePeriod(start, time.Hour, now)

	require.True(t, due)
	require.True(t, increments > 1000)
	require.True(t, nextStartedAt.Before(now) || nextStartedAt.Equal(now))
	require.True(t, nextStartedAt.Add(time.Hour).After(now))
}

func TestAdvancePeriodOneSecondInterval(t *testing.T) {
	t.Parallel()

	start := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	now := time.Date(2026, 3, 19, 12, 0, 5, 0, time.UTC)

	nextStartedAt, increments, due := advancePeriod(start, time.Second, now)

	require.True(t, due)
	require.EqualValues(t, 5, increments)
	require.Equal(t, time.Date(2026, 3, 19, 12, 0, 5, 0, time.UTC), nextStartedAt)
}
