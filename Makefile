APP_PACKAGE := ./cmd/api
MIGRATE_PACKAGE := ./cmd/migrate

.PHONY: run test test-race test-integration lint fmt docker-up migrate

run:
	go run $(APP_PACKAGE)

test:
	go test ./...

test-race:
	go test -race -count=1 -timeout=120s ./...

test-integration:
	go test -tags=integration -race -count=1 -timeout=300s ./...

lint:
	go vet ./...
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "Files not formatted:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

fmt:
	gofmt -w .

docker-up:
	docker compose up --build

migrate:
	go run $(MIGRATE_PACKAGE)
