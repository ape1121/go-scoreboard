package http

import "github.com/ape1121/go-scoreboard/internal/score"

type setScoreRequest struct {
	UserID string `json:"userId"`
	Score  int64  `json:"score"`
}

type setScoreResponse struct {
	BoardID string `json:"boardId"`
	UserID  string `json:"userId"`
	Score   int64  `json:"score"`
}

type scoreItemResponse struct {
	UserID string `json:"userId"`
	Score  int64  `json:"score"`
}

func (r setScoreRequest) toInput(boardID string) score.SetInput {
	return score.SetInput{
		BoardID: boardID,
		UserID:  r.UserID,
		Score:   r.Score,
	}
}

func toSetScoreResponse(entry score.ScoreEntry) setScoreResponse {
	return setScoreResponse{
		BoardID: entry.BoardID,
		UserID:  entry.UserID,
		Score:   entry.Score,
	}
}

func toTopScoresResponse(entries []score.ScoreEntry) []scoreItemResponse {
	response := make([]scoreItemResponse, 0, len(entries))
	for _, entry := range entries {
		response = append(response, scoreItemResponse{
			UserID: entry.UserID,
			Score:  entry.Score,
		})
	}

	return response
}

type seedRequest struct {
	Count    int   `json:"count"`
	MaxScore int64 `json:"maxScore"`
}

type seedResponse struct {
	Created int `json:"created"`
}

type surroundingsResponse struct {
	User  scoreItemResponse   `json:"user"`
	Above []scoreItemResponse `json:"above"`
	Below []scoreItemResponse `json:"below"`
}

func toSurroundingsResponse(entries []score.RankedEntry, targetUserID string) surroundingsResponse {
	var response surroundingsResponse
	response.Above = make([]scoreItemResponse, 0)
	response.Below = make([]scoreItemResponse, 0)

	phase := "above"
	for _, entry := range entries {
		item := scoreItemResponse{UserID: entry.UserID, Score: entry.Score}
		if entry.UserID == targetUserID {
			response.User = item
			phase = "below"
			continue
		}
		if phase == "above" {
			response.Above = append(response.Above, item)
		} else {
			response.Below = append(response.Below, item)
		}
	}

	return response
}
