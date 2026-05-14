package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/Yessenchik/order-service/domain"
)

type OrderRepository interface {
	Create(ctx context.Context, order domain.Order) error
	UpdateStatus(ctx context.Context, orderID, status string) error
	GetByID(ctx context.Context, orderID string) (*domain.Order, error)
	WaitForStatusChange(ctx context.Context, orderID string, lastUpdatedAt time.Time) (*domain.Order, error)
}

type PaymentClient interface {
	ProcessPayment(ctx context.Context, orderID string, amount float64, customerEmail string) (bool, string, error)
}

type OrderCache interface {
	Get(ctx context.Context, orderID string) (*domain.Order, bool)
	Set(ctx context.Context, order *domain.Order) error
	Delete(ctx context.Context, orderID string) error
}

type OrderUsecase struct {
	repo          OrderRepository
	paymentClient PaymentClient
	cache         OrderCache
}

func NewOrderUsecase(repo OrderRepository, paymentClient PaymentClient, cache OrderCache) *OrderUsecase {
	return &OrderUsecase{repo: repo, paymentClient: paymentClient, cache: cache}
}

func (u *OrderUsecase) CreateOrder(ctx context.Context, orderID string, amount float64, customerEmail string) (bool, string, error) {
	if orderID == "" {
		return false, "", errors.New("order id is required")
	}
	if amount <= 0 {
		return false, "", errors.New("amount must be greater than 0")
	}

	order := domain.Order{
		ID:     orderID,
		Amount: amount,
		Status: "PENDING",
	}
	if err := u.repo.Create(ctx, order); err != nil {
		return false, "", err
	}

	ok, msg, err := u.paymentClient.ProcessPayment(ctx, orderID, amount, customerEmail)
	if err != nil {
		_ = u.repo.UpdateStatus(ctx, orderID, "PAYMENT_FAILED")
		return false, "", err
	}

	if ok {
		_ = u.repo.UpdateStatus(ctx, orderID, "PAID")
	} else {
		_ = u.repo.UpdateStatus(ctx, orderID, "PAYMENT_FAILED")
	}
	if u.cache != nil {
		_ = u.cache.Delete(ctx, orderID)
	}

	return ok, msg, nil
}

func (u *OrderUsecase) UpdateOrderStatus(ctx context.Context, orderID, status string) error {
	if err := u.repo.UpdateStatus(ctx, orderID, status); err != nil {
		return err
	}
	if u.cache != nil {
		_ = u.cache.Delete(ctx, orderID)
	}
	return nil
}

func (u *OrderUsecase) GetOrder(ctx context.Context, orderID string) (*domain.Order, error) {
	if u.cache != nil {
		if order, ok := u.cache.Get(ctx, orderID); ok {
			return order, nil
		}
	}
	order, err := u.repo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if u.cache != nil {
		_ = u.cache.Set(ctx, order)
	}
	return order, nil
}
