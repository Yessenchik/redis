package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Yessenchik/payment-service/domain"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
)

type PaymentRepository interface {
	Create(ctx context.Context, p domain.Payment) error
	ListByStatus(ctx context.Context, status string) ([]domain.Payment, error)
}

type PaymentCompletedEvent struct {
	EventID       string  `json:"event_id"`
	OrderID       string  `json:"order_id"`
	Amount        float64 `json:"amount"`
	CustomerEmail string  `json:"customer_email"`
	Status        string  `json:"status"`
}

type PaymentUsecase struct {
	repo PaymentRepository
}

func NewPaymentUsecase(repo PaymentRepository) *PaymentUsecase {
	return &PaymentUsecase{repo: repo}
}

func (u *PaymentUsecase) ProcessPayment(ctx context.Context, orderID string, amount float64, customerEmail string) (bool, string, error) {
	if orderID == "" {
		return false, "", errors.New("order id is required")
	}

	if amount <= 0 {
		return false, "", errors.New("invalid payment amount")
	}

	if customerEmail == "" {
		return false, "", errors.New("customer email is required")
	}

	status := "COMPLETED"

	payment := domain.Payment{
		ID:      fmt.Sprintf("pay-%d", time.Now().UnixNano()),
		OrderID: orderID,
		Amount:  amount,
		Status:  status,
	}

	if err := u.repo.Create(ctx, payment); err != nil {
		return false, "", err
	}

	event := PaymentCompletedEvent{
		EventID:       uuid.New().String(),
		OrderID:       orderID,
		Amount:        amount,
		CustomerEmail: customerEmail,
		Status:        status,
	}

	if err := publishPaymentCompletedEvent(event); err != nil {
		return false, "", err
	}

	return true, "payment processed successfully", nil
}

func publishPaymentCompletedEvent(event PaymentCompletedEvent) error {
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@rabbitmq:5672/"
	}

	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(
		"payment.completed",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return ch.Publish(
		"",
		"payment.completed",
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func (u *PaymentUsecase) ListPayments(ctx context.Context, status string) ([]domain.Payment, error) {
	return u.repo.ListByStatus(ctx, status)
}
