package http

import (
	"time"

	"github.com/ape1121/go-scoreboard/internal/board"
)

type createBoardRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Schedule    *boardScheduleDTO `json:"schedule,omitempty"`
}

type boardScheduleDTO struct {
	Type            string `json:"type"`
	IntervalSeconds int64  `json:"intervalSeconds"`
}

type createBoardResponse struct {
	BoardID     string            `json:"boardId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Schedule    *boardScheduleDTO `json:"schedule,omitempty"`
}

type boardListItemResponse struct {
	BoardID string `json:"boardId"`
	Name    string `json:"name"`
}

type getBoardResponse struct {
	BoardID     string            `json:"boardId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	CreatedAt   time.Time         `json:"createdAt"`
	Schedule    *boardScheduleDTO `json:"schedule,omitempty"`
	NextResetAt *time.Time        `json:"nextResetAt,omitempty"`
}

func (r createBoardRequest) toInput() board.CreateInput {
	return board.CreateInput{
		Name:        r.Name,
		Description: r.Description,
		Schedule:    toDomainSchedule(r.Schedule),
	}
}

func toCreateBoardResponse(entity board.Board) createBoardResponse {
	return createBoardResponse{
		BoardID:     entity.ID,
		Name:        entity.Name,
		Description: entity.Description,
		Schedule:    toScheduleDTO(entity.Schedule),
	}
}

func toBoardListResponse(boards []board.Board) []boardListItemResponse {
	response := make([]boardListItemResponse, 0, len(boards))
	for _, entity := range boards {
		response = append(response, boardListItemResponse{
			BoardID: entity.ID,
			Name:    entity.Name,
		})
	}

	return response
}

func toGetBoardResponse(details board.Details) getBoardResponse {
	return getBoardResponse{
		BoardID:     details.Board.ID,
		Name:        details.Board.Name,
		Description: details.Board.Description,
		CreatedAt:   details.Board.CreatedAt,
		Schedule:    toScheduleDTO(details.Board.Schedule),
		NextResetAt: details.NextResetAt,
	}
}

func toDomainSchedule(schedule *boardScheduleDTO) *board.Schedule {
	if schedule == nil {
		return nil
	}

	return &board.Schedule{
		Type:     board.ScheduleType(schedule.Type),
		Interval: time.Duration(schedule.IntervalSeconds) * time.Second,
	}
}

func toScheduleDTO(schedule *board.Schedule) *boardScheduleDTO {
	if schedule == nil {
		return nil
	}

	return &boardScheduleDTO{
		Type:            string(schedule.Type),
		IntervalSeconds: schedule.IntervalSeconds(),
	}
}
