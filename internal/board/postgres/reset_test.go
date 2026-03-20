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
