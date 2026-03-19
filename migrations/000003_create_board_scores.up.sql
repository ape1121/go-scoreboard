CREATE TABLE board_scores (
    board_id TEXT NOT NULL,
    board_period_id BIGINT NOT NULL,
    user_id TEXT NOT NULL,
    score BIGINT NOT NULL,
    achieved_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (board_id, board_period_id, user_id),
    CONSTRAINT board_scores_period_fk
        FOREIGN KEY (board_id, board_period_id)
        REFERENCES board_periods (board_id, period_id)
        ON DELETE CASCADE,
    CONSTRAINT board_scores_user_id_not_blank CHECK (BTRIM(user_id) <> ''),
    CONSTRAINT board_scores_score_non_negative CHECK (score >= 0)
);

CREATE INDEX board_scores_leaderboard_rank_idx
    ON board_scores (board_id, board_period_id, score DESC, achieved_at ASC, user_id ASC);
