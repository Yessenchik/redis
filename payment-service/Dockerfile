FROM golang:1.25 AS builder

WORKDIR /app

COPY generated ./generated
COPY payment-service/go.mod payment-service/go.sum ./payment-service/

WORKDIR /app/payment-service
RUN go mod download

COPY payment-service ./
RUN go build -o payment-service ./cmd/

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/payment-service/payment-service .
EXPOSE 50051
CMD ["./payment-service"]