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
)

func TestCreateBoardHandlerReturnsCreatedBoard(t *testing.T) {
	t.Parallel()

	router := newBoardTestRouter(&boardRepositoryStub{}, time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBufferString(`{
		"name":"Weekly Tournament",
		"description":"Global leaderboard for weekly tournament",
		"schedule":{"type":"interval","intervalSeconds":604800}
	}`))

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusCreated, recorder.Code)
	require.JSONEq(t, `{
		"boardId":"board_test",
		"name":"Weekly Tournament",
		"description":"Global leaderboard for weekly tournament",
		"schedule":{"type":"interval","intervalSeconds":604800}
	}`, recorder.Body.String())
}

func TestCreateBoardHandlerReturnsBadRequestForInvalidPayload(t *testing.T) {
	t.Parallel()

	router := newBoardTestRouter(&boardRepositoryStub{}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBufferString(`{"name":"","description":"x"}`))

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)
	require.JSONEq(t, `{"error":"name must not be empty"}`, recorder.Body.String())
}

func TestListBoardsHandlerReturnsBoardSummaries(t *testing.T) {
	t.Parallel()

	router := newBoardTestRouter(&boardRepositoryStub{
		listBoards: []board.Board{
			{ID: "board_1", Name: "Weekly Tournament"},
			{ID: "board_2", Name: "All-time Top Scores"},
		},
	}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `[
		{"boardId":"board_1","name":"Weekly Tournament"},
		{"boardId":"board_2","name":"All-time Top Scores"}
	]`, recorder.Body.String())
}

func TestGetBoardHandlerReturnsBoardDetails(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC)
	router := newBoardTestRouter(&boardRepositoryStub{
		boardByID: board.Board{
			ID:          "board_test",
			Name:        "Weekly Tournament",
			Description: "Global leaderboard for weekly tournament",
			CreatedAt:   startedAt,
			Schedule: &board.Schedule{
				Type:     board.ScheduleTypeInterval,
				Interval: 7 * 24 * time.Hour,
			},
		},
		activePeriod: board.BoardPeriod{
			ID:        1,
			BoardID:   "board_test",
			Sequence:  0,
			StartedAt: startedAt,
		},
	}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards/board_test", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `{
		"boardId":"board_test",
		"name":"Weekly Tournament",
		"description":"Global leaderboard for weekly tournament",
		"createdAt":"2026-03-19T12:00:00Z",
		"schedule":{"type":"interval","intervalSeconds":604800},
		"nextResetAt":"2026-03-26T12:00:00Z"
	}`, recorder.Body.String())
}

func TestGetBoardHandlerReturnsNotFound(t *testing.T) {
	t.Parallel()

	router := newBoardTestRouter(&boardRepositoryStub{getErr: board.ErrNotFound}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards/board_missing", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)
	require.JSONEq(t, `{"error":"Board not found"}`, recorder.Body.String())
}

func TestListBoardsHandlerAcceptsPaginationParams(t *testing.T) {
	t.Parallel()

	router := newBoardTestRouter(&boardRepositoryStub{
		listBoards: []board.Board{
			{ID: "board_3", Name: "Third"},
		},
	}, time.Now().UTC())
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/boards?limit=1&offset=2", nil)

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	require.JSONEq(t, `[{"boardId":"board_3","name":"Third"}]`, recorder.Body.String())
}

func TestCreateAndListBoardsRoundTrip(t *testing.T) {
	t.Parallel()

	repo := &boardRepositoryStub{}
	router := newBoardTestRouter(repo, time.Date(2026, 3, 19, 12, 0, 0, 0, time.UTC))

	createRecorder := httptest.NewRecorder()
	createReq := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBufferString(`{
		"name":"Round Trip Board",
		"description":"Testing create then list"
	}`))
	router.ServeHTTP(createRecorder, createReq)
	require.Equal(t, http.StatusCreated, createRecorder.Code)

	listRecorder := httptest.NewRecorder()
	listReq := httptest.NewRequest(http.MethodGet, "/boards", nil)
	router.ServeHTTP(listRecorder, listReq)
	require.Equal(t, http.StatusOK, listRecorder.Code)
	require.Contains(t, listRecorder.Body.String(), "Round Trip Board")
}

func newBoardTestRouter(repository *boardRepositoryStub, now time.Time) http.Handler {
	service := board.NewService(repository, fixedClock{now: now}, func() string { return "board_test" })

	return NewRouter(Dependencies{
		Logger:        log.New(&bytes.Buffer{}, "", 0),
		HealthService: NewHealthService(stubPinger{}),
		BoardService:  service,
	})
}

type boardRepositoryStub struct {
	createdBoard  board.Board
	createdPeriod board.BoardPeriod
	listBoards    []board.Board
	boardByID     board.Board
	activePeriod  board.BoardPeriod
	createErr     error
	listErr       error
	getErr        error
	periodErr     error
}

func (s *boardRepositoryStub) Create(_ context.Context, boardEntity board.Board, period board.BoardPeriod) error {
	if s.createErr != nil {
		return s.createErr
	}

	s.createdBoard = boardEntity
	s.createdPeriod = period
	if s.boardByID.ID == "" {
		s.boardByID = boardEntity
	}
	if s.activePeriod.BoardID == "" {
		s.activePeriod = period
	}
	return nil
}

func (s *boardRepositoryStub) List(_ context.Context, _, _ int) ([]board.Board, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}

	if s.listBoards != nil {
		return s.listBoards, nil
	}
	if s.createdBoard.ID != "" {
		return []board.Board{{ID: s.createdBoard.ID, Name: s.createdBoard.Name}}, nil
	}
	return nil, nil
}

func (s *boardRepositoryStub) GetByID(_ context.Context, _ string) (board.Board, error) {
	if s.getErr != nil {
		return board.Board{}, s.getErr
	}

	return s.boardByID, nil
}

func (s *boardRepositoryStub) GetActivePeriod(_ context.Context, _ string) (board.BoardPeriod, error) {
	if s.periodErr != nil {
		return board.BoardPeriod{}, s.periodErr
	}

	return s.activePeriod, nil
}

type fixedClock struct {
	now time.Time
}

func (f fixedClock) Now() time.Time {
	return f.now
}
