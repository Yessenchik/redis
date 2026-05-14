package postgres

import (
	"context"

	"github.com/Yessenchik/payment-service/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepository struct {
	db *pgxpool.Pool
}

func NewPaymentRepository(db *pgxpool.Pool) *PaymentRepository {
	return &PaymentRepository{db: db}
}

func (r *PaymentRepository) Create(ctx context.Context, p domain.Payment) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO payments (id, order_id, amount, status, created_at)
		VALUES ($1, $2, $3, $4, NOW())
	`, p.ID, p.OrderID, p.Amount, p.Status)
	return err
}

func (r *PaymentRepository) ListByStatus(ctx context.Context, status string) ([]domain.Payment, error) {
	var rows interface {
		Close()
		Next() bool
		Scan(dest ...any) error
		Err() error
	}

	var err error
	if status == "" {
		rows, err = r.db.Query(ctx, `
			SELECT id, order_id, amount, status, created_at
			FROM payments
			ORDER BY created_at DESC
		`)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT id, order_id, amount, status, created_at
			FROM payments
			WHERE status = $1
			ORDER BY created_at DESC
		`, status)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []domain.Payment
	for rows.Next() {
		var p domain.Payment
		if err := rows.Scan(&p.ID, &p.OrderID, &p.Amount, &p.Status, &p.CreatedAt); err != nil {
			return nil, err
		}
		payments = append(payments, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return payments, nil
}
