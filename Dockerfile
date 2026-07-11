FROM golang:1.26.2-alpine AS builder

WORKDIR /application

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /application/ecommerce ./cmd/api

FROM alpine:3.23

WORKDIR /application

RUN addgroup -S application && adduser -S application -G application

COPY --from=builder /application/ecommerce ./ecommerce

USER application

EXPOSE 3000

CMD ["./ecommerce"]
