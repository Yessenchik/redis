package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type PaymentCompletedEvent struct {
	EventID       string  `json:"event_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
}

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type SimulatedEmailSender struct{}

func (s *SimulatedEmailSender) Send(ctx context.Context, to, subject, body string) error {
	time.Sleep(1500 * time.Millisecond)
	if rand.Intn(100) < 80 {
		return errors.New("simulated provider temporary failure")
	}
	log.Printf("[Provider] Sent email to=%s subject=%s body=%s", to, subject, body)
	return nil
}

func NewEmailSender() EmailSender {
	mode := os.Getenv("PROVIDER_MODE")
	if mode == "REAL" {
		// Real SMTP/Mailjet adapter can be plugged in here without changing worker logic.
		log.Println("PROVIDER_MODE=REAL selected, but SMTP adapter is not configured; using simulated provider")
	}
	return &SimulatedEmailSender{}
}

func getEnv(key, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func processNotification(ctx context.Context, rdb *redis.Client, sender EmailSender, event PaymentCompletedEvent, maxRetries int) error {
	idempotencyKey := "notification:payment:" + event.EventID

	alreadyProcessed, err := rdb.Exists(ctx, idempotencyKey).Result()
	if err == nil && alreadyProcessed == 1 {
		log.Println("Duplicate notification skipped:", event.EventID)
		return nil
	}

	subject := "Payment completed"
	body := "Your payment for order " + event.OrderID + " was completed."

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = sender.Send(ctx, event.CustomerEmail, subject, body)
		if err == nil {
			return rdb.Set(ctx, idempotencyKey, "sent", 24*time.Hour).Err()
		}

		backoff := time.Duration(1<<attempt) * time.Second // 2s, 4s, 8s...
		log.Printf("Notification failed event_id=%s attempt=%d error=%v retry_in=%s", event.EventID, attempt, err, backoff)
		time.Sleep(backoff)
	}

	_ = rdb.Set(ctx, idempotencyKey, "failed", 24*time.Hour).Err()
	return errors.New("notification failed after retries")
}

func main() {
	url := getEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	redisAddr := getEnv("REDIS_ADDR", "redis:6379")
	maxRetries := getEnvInt("MAX_RETRIES", 3)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()

	sender := NewEmailSender()

	var conn *amqp.Connection
	var err error
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(url)
		if err == nil {
			log.Println("Connected to RabbitMQ")
			break
		}
		log.Println("RabbitMQ not ready, retrying...")
		time.Sleep(3 * time.Second)
	}
	if err != nil {
		log.Fatal("Failed to connect to RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	queue, err := ch.QueueDeclare("payment.completed", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Notification Service worker started")
	for {
		select {
		case <-ctx.Done():
			log.Println("Shutting down Notification Service...")
			return
		case msg := <-msgs:
			var event PaymentCompletedEvent
			if err := json.Unmarshal(msg.Body, &event); err != nil {
				log.Println("Invalid message:", err)
				msg.Nack(false, false)
				continue
			}

			if err := processNotification(ctx, rdb, sender, event, maxRetries); err != nil {
				log.Println(err)
				msg.Nack(false, false)
				continue
			}
			msg.Ack(false)
		}
	}
}
