FROM golang:1.25 AS builder

WORKDIR /app

COPY generated ./generated

COPY order-service/go.mod order-service/go.sum ./order-service/

WORKDIR /app/order-service
RUN go mod download

# copy source code
COPY order-service ./

RUN go build -o order-service ./cmd/

FROM debian:bookworm-slim
WORKDIR /app
COPY --from=builder /app/order-service/order-service .
EXPOSE 8080 50052
CMD ["./order-service"]