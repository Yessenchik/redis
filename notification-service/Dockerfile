FROM golang:1.25 AS builder

WORKDIR /app

COPY generated ./generated
COPY notification-service/go.mod notification-service/go.sum ./notification-service/

WORKDIR /app/notification-service
RUN go mod download

COPY notification-service ./
RUN go build -o notification-service .

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/notification-service/notification-service .
CMD ["./notification-service"]