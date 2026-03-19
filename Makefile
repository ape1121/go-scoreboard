APP_PACKAGE := ./cmd/api
MIGRATE_PACKAGE := ./cmd/migrate

.PHONY: run test lint docker-up migrate

run:
	go run $(APP_PACKAGE)

test:
	go test ./...

lint:
	go vet ./...

docker-up:
	docker compose up --build

migrate:
	go run $(MIGRATE_PACKAGE)
