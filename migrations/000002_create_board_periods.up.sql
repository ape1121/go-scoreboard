CREATE TABLE board_periods (
    period_id BIGSERIAL PRIMARY KEY,
    board_id TEXT NOT NULL,
    sequence_number BIGINT NOT NULL,
    started_at TIMESTAMPTZ NOT NULL,
    ended_at TIMESTAMPTZ,
    CONSTRAINT board_periods_board_fk
        FOREIGN KEY (board_id) REFERENCES boards (board_id) ON DELETE CASCADE,
    CONSTRAINT board_periods_board_sequence_unique
        UNIQUE (board_id, sequence_number),
    CONSTRAINT board_periods_board_period_unique
        UNIQUE (board_id, period_id),
    CONSTRAINT board_periods_time_window_valid
        CHECK (ended_at IS NULL OR ended_at > started_at)
);

CREATE UNIQUE INDEX board_periods_active_lookup_idx
    ON board_periods (board_id)
    WHERE ended_at IS NULL;
