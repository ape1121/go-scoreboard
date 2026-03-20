#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8080}"

echo "=== Health Check ==="
curl -s "$BASE_URL/healthz" | jq .

echo ""
echo "=== Create Board (with schedule) ==="
BOARD=$(curl -s -X POST "$BASE_URL/boards" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Weekly Tournament",
    "description": "Global leaderboard for weekly tournament",
    "schedule": {
      "type": "interval",
      "intervalSeconds": 604800
    }
  }')
echo "$BOARD" | jq .
BOARD_ID=$(echo "$BOARD" | jq -r '.boardId')

echo ""
echo "=== Create Board (no schedule) ==="
curl -s -X POST "$BASE_URL/boards" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "All-time Top Scores",
    "description": "Persistent leaderboard"
  }' | jq .

echo ""
echo "=== List Boards ==="
curl -s "$BASE_URL/boards" | jq .

echo ""
echo "=== List Boards (with pagination) ==="
curl -s "$BASE_URL/boards?limit=1&offset=0" | jq .

echo ""
echo "=== Get Board ==="
curl -s "$BASE_URL/boards/$BOARD_ID" | jq .

echo ""
echo "=== Set Scores ==="
curl -s -X POST "$BASE_URL/boards/$BOARD_ID/scores" \
  -H "Content-Type: application/json" \
  -d '{"userId": "alice", "score": 2500}' | jq .

curl -s -X POST "$BASE_URL/boards/$BOARD_ID/scores" \
  -H "Content-Type: application/json" \
  -d '{"userId": "bob", "score": 3200}' | jq .

curl -s -X POST "$BASE_URL/boards/$BOARD_ID/scores" \
  -H "Content-Type: application/json" \
  -d '{"userId": "carol", "score": 1800}' | jq .

echo ""
echo "=== Get Top Scores ==="
curl -s "$BASE_URL/boards/$BOARD_ID/scores?n=10" | jq .

echo ""
echo "=== Seed Board with Mock Data ==="
curl -s -X POST "$BASE_URL/boards/$BOARD_ID/scores/seed" \
  -H "Content-Type: application/json" \
  -d '{"count": 10, "maxScore": 5000}' | jq .

echo ""
echo "=== Get Top Scores (after seed) ==="
curl -s "$BASE_URL/boards/$BOARD_ID/scores?n=5" | jq .

echo ""
echo "=== Get Score Surroundings ==="
curl -s "$BASE_URL/boards/$BOARD_ID/scores/alice/surroundings?n=2" | jq .

echo ""
echo "=== Error: Board Not Found ==="
curl -s "$BASE_URL/boards/nonexistent" | jq .

echo ""
echo "=== Error: Invalid Score ==="
curl -s -X POST "$BASE_URL/boards/$BOARD_ID/scores" \
  -H "Content-Type: application/json" \
  -d '{"userId": "", "score": -1}' | jq .
