package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/Yessenchik/order-service/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order domain.Order) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO orders (id, amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
	`, order.ID, order.Amount, order.Status)
	return err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID, status string) error {
	cmd, err := r.db.Exec(ctx, `
		UPDATE orders
		SET status = $2, updated_at = NOW()
		WHERE id = $1
	`, orderID, status)
	if err != nil {
		return err
	}
	if cmd.RowsAffected() == 0 {
		return errors.New("order not found")
	}
	return nil
}

func (r *OrderRepository) GetByID(ctx context.Context, orderID string) (*domain.Order, error) {
	row := r.db.QueryRow(ctx, `
		SELECT id, amount, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`, orderID)

	var o domain.Order
	if err := row.Scan(&o.ID, &o.Amount, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) WaitForStatusChange(
	ctx context.Context,
	orderID string,
	lastUpdatedAt time.Time,
) (*domain.Order, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			order, err := r.GetByID(ctx, orderID)
			if err != nil {
				return nil, err
			}
			if order.UpdatedAt.After(lastUpdatedAt) {
				return order, nil
			}
		}
	}
}
