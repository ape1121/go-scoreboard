package http

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/ape1121/go-scoreboard/internal/board"
	"github.com/ape1121/go-scoreboard/internal/score"
)

func TestSetScoreHandlerReturnsScore(t *testing.T) {
	t.Parallel()

	router := newScoreTestRouter(&scoreRepositoryStub{}, &scoreBoardResolverStub{
		boardEntity: board.Board{ID: "board_test"},
		period:      board.BoardPeriod{ID: 11, BoardID: "board_test"},
	}, time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/boards/board_test/scores", bytes.NewBufferString(`{"userId":"user_1","score":1500}`))

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{"boardId":"board_test","userId":"user_1","score":1500}`, recorder.Body.String())
}

func TestSetScoreHandlerReturnsBoardNotFound(t *testing.T) {
	t.Parallel()

	router := newScoreTestRouter(&scoreRepositoryStub{}, &scoreBoardResolverStub{getErr: board.ErrNotFound}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/boards/board_test/scores", bytes.NewBufferString(`{"userId":"user_1","score":1500}`))

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.JSONEq(t, `{"error":"board not found"}`, recorder.Body.String())
}

func TestTopScoresHandlerReturnsRankedEntries(t *testing.T) {
	t.Parallel()

	router := newScoreTestRouter(&scoreRepositoryStub{
		topEntries: []score.ScoreEntry{
			{BoardID: "board_test", PeriodID: 11, UserID: "user_2", Score: 2000},
			{BoardID: "board_test", PeriodID: 11, UserID: "user_1", Score: 1500},
		},
	}, &scoreBoardResolverStub{
		boardEntity: board.Board{ID: "board_test"},
		period:      board.BoardPeriod{ID: 11, BoardID: "board_test"},
	}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards/board_test/scores?n=2", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `[
		{"userId":"user_2","score":2000},
		{"userId":"user_1","score":1500}
	]`, recorder.Body.String())
}

func TestTopScoresHandlerReturnsBadRequestForInvalidLimit(t *testing.T) {
	t.Parallel()

	router := newScoreTestRouter(&scoreRepositoryStub{}, &scoreBoardResolverStub{
		boardEntity: board.Board{ID: "board_test"},
		period:      board.BoardPeriod{ID: 11, BoardID: "board_test"},
	}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards/board_test/scores?n=0", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.JSONEq(t, `{"error":"limit must be at least 1"}`, recorder.Body.String())
}

func TestTopScoresHandlerReturnsEmptyArrayWhenNoScoresExist(t *testing.T) {
	t.Parallel()

	router := newScoreTestRouter(&scoreRepositoryStub{}, &scoreBoardResolverStub{
		boardEntity: board.Board{ID: "board_test"},
		period:      board.BoardPeriod{ID: 11, BoardID: "board_test"},
	}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards/board_test/scores", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `[]`, recorder.Body.String())
}

func newScoreTestRouter(repository *scoreRepositoryStub, boards *scoreBoardResolverStub, now time.Time) http.Handler {
	service := score.NewService(repository, boards, fixedScoreClock{now: now})

	return NewRouter(Dependencies{
		Logger:        log.New(&bytes.Buffer{}, "", 0),
		HealthService: NewHealthService(stubPinger{}),
		ScoreService:  service,
	})
}

type scoreRepositoryStub struct {
	upserted   score.ScoreEntry
	topEntries []score.ScoreEntry
	upsertErr  error
	topErr     error
}

func (s *scoreRepositoryStub) Upsert(_ context.Context, entry score.ScoreEntry) error {
	if s.upsertErr != nil {
		return s.upsertErr
	}

	s.upserted = entry
	return nil
}

func (s *scoreRepositoryStub) Top(_ context.Context, _ string, _ int64, _ int) ([]score.ScoreEntry, error) {
	if s.topErr != nil {
		return nil, s.topErr
	}

	return s.topEntries, nil
}

func (s *scoreRepositoryStub) Get(context.Context, string, int64, string) (score.ScoreEntry, error) {
	return score.ScoreEntry{}, nil
}

func (s *scoreRepositoryStub) Surroundings(context.Context, string, int64, string, int) ([]score.ScoreEntry, []score.ScoreEntry, score.ScoreEntry, error) {
	return nil, nil, score.ScoreEntry{}, nil
}

type scoreBoardResolverStub struct {
	boardEntity board.Board
	period      board.BoardPeriod
	getErr      error
	periodErr   error
}

func (s *scoreBoardResolverStub) GetByID(context.Context, string) (board.Board, error) {
	if s.getErr != nil {
		return board.Board{}, s.getErr
	}

	return s.boardEntity, nil
}

func (s *scoreBoardResolverStub) GetActivePeriod(context.Context, string) (board.BoardPeriod, error) {
	if s.periodErr != nil {
		return board.BoardPeriod{}, s.periodErr
	}

	return s.period, nil
}

type fixedScoreClock struct {
	now time.Time
}

func (f fixedScoreClock) Now() time.Time {
	return f.now
}
