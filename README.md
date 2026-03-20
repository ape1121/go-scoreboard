# Go Scoreboard

REST API for managing leaderboards with optional periodic score resets.

## Quick Start

```bash
docker compose up --build
```

The API starts on `http://localhost:8080`. Migrations are applied automatically on startup.

Run the curl examples to verify:

```bash
bash scripts/curl-examples.sh
```

## API

### Boards

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/boards` | Create a board |
| `GET` | `/boards` | List boards (`?limit=50&offset=0`) |
| `GET` | `/boards/{boardId}` | Get board details |

### Scores

| Method | Path | Description |
|--------|------|-------------|
| `POST` | `/boards/{boardId}/scores` | Set a user's score |
| `GET` | `/boards/{boardId}/scores?n=10` | Get top scores |
| `GET` | `/boards/{boardId}/scores/{userId}/surroundings?n=5` | Scores around a user |
| `POST` | `/boards/{boardId}/scores/seed` | Seed mock scores |

### Health

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/healthz` | Health check |

## Examples

Create a board:

```bash
curl -X POST http://localhost:8080/boards \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekly Tournament",
    "description": "Resets every week",
    "schedule": {"type": "interval", "intervalSeconds": 604800}
  }'
```

Set a score:

```bash
curl -X POST http://localhost:8080/boards/{boardId}/scores \
  -H "Content-Type: application/json" \
  -d '{"userId": "alice", "score": 2500}'
```

Get leaderboard:

```bash
curl http://localhost:8080/boards/{boardId}/scores?n=10
```

Seed test data:

```bash
curl -X POST http://localhost:8080/boards/{boardId}/scores/seed \
  -H "Content-Type: application/json" \
  -d '{"count": 50, "maxScore": 10000}'
```

## Local Development

Requirements: Go 1.22+, Docker

```bash
# Start Postgres
docker compose up postgres -d

# Run migrations and start the API
make migrate
make run
```

## Testing

```bash
# Unit tests
make test

# Unit tests with race detector (requires GCC / Linux / macOS)
make test-race

# Integration tests (requires Docker)
make test-integration

# Lint
make lint
```

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `HTTP_PORT` | `8080` | HTTP listen port |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/scoreboard?sslmode=disable` | Postgres connection string |
| `SCHEDULER_POLL_INTERVAL` | `10s` | How often to check for due resets |
| `SHUTDOWN_TIMEOUT` | `10s` | Graceful shutdown timeout |

## Architecture

### Key design choices

- **Modular monolith** with clear boundaries between HTTP, domain, and persistence layers.
- **SQL-first persistence** with explicit queries instead of an ORM, giving full control over ranking and transaction behavior.
- **Board periods** model resets explicitly: each reset closes the current period, deletes its scores, and opens a new one transactionally.
- **Database-level concurrency control** via `FOR UPDATE` locks ensures score writes and resets serialize correctly per board.
- **In-process scheduler** polls for due boards and catches up any missed resets on startup.

### Tradeoffs and future improvements

- Offset-based pagination is simple but O(n) for deep pages; cursor-based pagination would scale better.
- The scheduler is in-process; a distributed lock (e.g., advisory locks) would be needed for multi-instance deployments.
- No authentication or rate limiting; these would be necessary for production use.
- Historical scores are permanently deleted on reset; archiving to a separate table could be added if needed.
