SQLC_VERSION := 1.31.1

.PHONY: run build test coverage vet check compose-up compose-down database-up migrate-up migrate-down sqlc sqlc-check demo

run:
	go run ./cmd/api

build:
	go build -o ./bin/ecommerce ./cmd/api

test:
	go test -count=1 ./...

coverage:
	./scripts/coverage.sh

vet:
	go vet ./...

check: sqlc-check
	gofmt -w cmd internal
	go test -count=1 ./...
	go vet ./...
	go build ./...

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

database-up:
	docker compose up -d postgresql

migrate-up:
	go run ./cmd/migrate -direction up

migrate-down:
	go run ./cmd/migrate -direction down

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v$(SQLC_VERSION) generate

sqlc-check: sqlc
	git diff --exit-code -- internal/platform/database/sqlc

demo:
	./scripts/http-flow.sh
