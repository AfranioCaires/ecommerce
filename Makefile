.PHONY: run build test vet check compose-up compose-down database-up swagger

run:
	go run ./cmd/api

build:
	go build -o ./bin/ecommerce ./cmd/api

test:
	go test -count=1 ./...

vet:
	go vet ./...

check:
	gofmt -w cmd internal
	go test -count=1 ./...
	go vet ./...

compose-up:
	docker compose up --build -d

compose-down:
	docker compose down

database-up:
	docker compose up -d postgresql

swagger:
	go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/api/main.go -o docs/swagger --parseInternal
