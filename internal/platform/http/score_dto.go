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

type rankedScoreItemResponse struct {
	Rank   int    `json:"rank"`
	UserID string `json:"userId"`
	Score  int64  `json:"score"`
}

func toSurroundingsResponse(entries []score.RankedEntry) []rankedScoreItemResponse {
	response := make([]rankedScoreItemResponse, 0, len(entries))
	for _, entry := range entries {
		response = append(response, rankedScoreItemResponse{
			Rank:   entry.Rank,
			UserID: entry.UserID,
			Score:  entry.Score,
		})
	}

	return response
}
