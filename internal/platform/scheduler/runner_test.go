package scheduler

import (
	"bytes"
	"context"
	"errors"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDBRunnerResetsDueBoards(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{dueBoardIDs: []string{"board_1", "board_2"}}
	runner := NewRunner(repository)
	now := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)

	err := runner.Run(context.Background(), now)

	require.NoError(t, err)
	require.Equal(t, []resetCall{
		{boardID: "board_1", now: now},
		{boardID: "board_2", now: now},
	}, repository.resetCalls)
}

func TestDBRunnerSkipsBoardsWithoutSchedulesWhenNoneAreDue(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{}
	runner := NewRunner(repository)

	err := runner.Run(context.Background(), time.Now().UTC())

	require.NoError(t, err)
	require.Empty(t, repository.resetCalls)
}

func TestSchedulerCatchUpUsesCurrentClock(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 3, 19, 13, 0, 0, 0, time.UTC)
	runner := &runnerStub{}
	scheduler := New(log.New(&bytes.Buffer{}, "", 0), time.Second, fixedClock{now: now}, runner)

	err := scheduler.CatchUp(context.Background())

	require.NoError(t, err)
	require.Equal(t, []time.Time{now}, runner.getCalls())
}

func TestSchedulerStartTriggersRunner(t *testing.T) {
	t.Parallel()

	runner := &runnerStub{}
	scheduler := New(log.New(&bytes.Buffer{}, "", 0), 10*time.Millisecond, fixedClock{now: time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)}, runner)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	scheduler.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	time.Sleep(50 * time.Millisecond)

	require.NotEmpty(t, runner.getCalls())
}

type repositoryStub struct {
	dueBoardIDs []string
	dueErr      error
	resetErr    error
	resetCalls  []resetCall
}

func (s *repositoryStub) DueBoardIDs(context.Context, time.Time) ([]string, error) {
	if s.dueErr != nil {
		return nil, s.dueErr
	}

	return s.dueBoardIDs, nil
}

func (s *repositoryStub) ResetDueBoard(_ context.Context, boardID string, now time.Time) (bool, error) {
	s.resetCalls = append(s.resetCalls, resetCall{boardID: boardID, now: now})
	if s.resetErr != nil {
		return false, s.resetErr
	}

	return true, nil
}

type resetCall struct {
	boardID string
	now     time.Time
}

type runnerStub struct {
	mu    sync.Mutex
	calls []time.Time
	err   error
}

func (s *runnerStub) Run(_ context.Context, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, now)
	return s.err
}

func (s *runnerStub) getCalls() []time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]time.Time{}, s.calls...)
}

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}

func TestDBRunnerReturnsResetErrors(t *testing.T) {
	t.Parallel()

	repository := &repositoryStub{
		dueBoardIDs: []string{"board_1"},
		resetErr:    errors.New("boom"),
	}
	runner := NewRunner(repository)

	err := runner.Run(context.Background(), time.Now().UTC())

	require.Error(t, err)
	require.Contains(t, err.Error(), "board_1")
}
