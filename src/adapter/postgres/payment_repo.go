package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"backend/src/domain/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PaymentRepositoryPGX struct{ pool *pgxpool.Pool }

func NewPaymentRepository(pool *pgxpool.Pool) *PaymentRepositoryPGX {
	return &PaymentRepositoryPGX{pool: pool}
}

func (r *PaymentRepositoryPGX) Pay(ctx context.Context, orderID, memberID, method, last4, masked string, now time.Time) (*model.Payment, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var owner, status, paymentStatus string
	var amount string
	err = tx.QueryRow(ctx, `SELECT customer_id,status,payment_status,total_amount::text FROM orders WHERE id=$1 FOR UPDATE`, orderID).Scan(&owner, &status, &paymentStatus, &amount)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if owner != memberID {
		return nil, model.ErrOrderForbidden
	}
	if status != "pending" || paymentStatus != "pending" {
		return nil, model.ErrPaymentNotAllowed
	}
	var payment model.Payment
	err = tx.QueryRow(ctx, `INSERT INTO payments(order_id,member_id,method,status,amount,masked_card,last4,created_at) VALUES($1,$2,$3,'paid',$4,$5,$6,$7) RETURNING id,created_at`, orderID, memberID, method, amount, masked, last4, now).Scan(&payment.ID, &payment.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, model.ErrPaymentAlreadyExists
		}
		return nil, fmt.Errorf("create payment: %w", err)
	}
	if _, err = tx.Exec(ctx, `UPDATE orders SET payment_status='paid',status='confirmed',updated_at=$2 WHERE id=$1`, orderID, now); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO order_status_history(order_id,status,changed_by,created_at) VALUES($1,'confirmed',$2,$3)`, orderID, memberID, now); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	payment.OrderID, payment.MemberID, payment.Method, payment.Status, payment.Amount, payment.MaskedCard, payment.Last4 = orderID, memberID, method, "paid", amount, masked, last4
	return &payment, nil
}

func (r *PaymentRepositoryPGX) ExpirePending(ctx context.Context, cutoff, now time.Time) (int, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `UPDATE orders SET status='cancelled',payment_status='failed',fulfillment_status='cancelled',updated_at=$2 WHERE status='pending' AND payment_status='pending' AND created_at <= $1 RETURNING id`, cutoff, now)
	if err != nil {
		return 0, err
	}
	ids, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) { var id string; return id, row.Scan(&id) })
	if err != nil {
		return 0, err
	}
	for _, id := range ids {
		if _, err = tx.Exec(ctx, `UPDATE inventory_reservations SET status='released',released_at=$2 WHERE order_id=$1 AND status='active'`, id, now); err != nil {
			return 0, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO order_status_history(order_id,status,created_at) VALUES($1,'cancelled',$2)`, id, now); err != nil {
			return 0, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(ids), nil
}
