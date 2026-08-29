SQLC_VERSION := 1.31.1

.PHONY: run run-api run-payment build build-api build-payment test coverage vet check compose-up compose-down database-up payment-database-up migrate-up migrate-down sqlc sqlc-check demo

run: run-api

run-api:
	go run ./cmd/api

run-payment:
	go run ./cmd/payment

build: build-api build-payment

build-api:
	mkdir -p ./bin
	go build -o ./bin/ecommerce-api ./cmd/api

build-payment:
	mkdir -p ./bin
	go build -o ./bin/payment-service ./cmd/payment

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
	docker compose up -d ecommerce-postgresql rabbitmq

payment-database-up:
	docker compose up -d payment-postgresql rabbitmq

migrate-up:
	go run ./cmd/migrate -direction up

migrate-down:
	go run ./cmd/migrate -direction down

sqlc:
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v$(SQLC_VERSION) generate
	go run github.com/sqlc-dev/sqlc/cmd/sqlc@v$(SQLC_VERSION) -f sqlc-payment.yaml generate

sqlc-check: sqlc
	git diff --exit-code -- internal/platform/database/sqlc internal/payment/platform/database/sqlc

demo:
	./scripts/http-flow.sh
