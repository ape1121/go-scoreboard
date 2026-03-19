CREATE TABLE boards (
    board_id TEXT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description VARCHAR(500) NOT NULL DEFAULT '',
    schedule_type TEXT,
    schedule_interval_seconds BIGINT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT boards_name_not_blank CHECK (BTRIM(name) <> ''),
    CONSTRAINT boards_schedule_type_valid CHECK (
        schedule_type IS NULL OR schedule_type = 'interval'
    ),
    CONSTRAINT boards_schedule_pairing_valid CHECK (
        (schedule_type IS NULL AND schedule_interval_seconds IS NULL)
        OR
        (schedule_type = 'interval' AND schedule_interval_seconds IS NOT NULL)
    ),
    CONSTRAINT boards_schedule_interval_valid CHECK (
        schedule_interval_seconds IS NULL
        OR schedule_interval_seconds BETWEEN 1 AND 31536000
    )
);
