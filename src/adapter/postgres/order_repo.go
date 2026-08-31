package postgres

import (
	"context"
	"errors"
	"fmt"

	"backend/src/domain/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepositoryPGX struct{ pool *pgxpool.Pool }

func NewOrderRepositoryPGX(pool *pgxpool.Pool) *OrderRepositoryPGX {
	return &OrderRepositoryPGX{pool: pool}
}

func (r *OrderRepositoryPGX) Create(ctx context.Context, order *model.Order, items []model.OrderItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if order.OrderNo == "" {
		order.OrderNo = "ORD-" + uuid.NewString()
	}
	if len(order.CustomerSnapshot) == 0 {
		order.CustomerSnapshot = []byte(`{}`)
	}
	err = tx.QueryRow(ctx, `INSERT INTO orders(order_no, customer_id, customer_snapshot, shipping_address_snapshot, subtotal, total_amount)
		VALUES($1,$2,$3::jsonb,$4::jsonb,0,0) RETURNING id, status, payment_status, fulfillment_status, created_at, updated_at`,
		order.OrderNo, order.CustomerID, order.CustomerSnapshot, order.ShippingAddressSnapshot).
		Scan(&order.ID, &order.Status, &order.PaymentStatus, &order.FulfillmentStatus, &order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create order: %w", err)
	}
	var subtotal float64
	for i := range items {
		var price float64
		var snapshot []byte
		var inventoryID string
		err = tx.QueryRow(ctx, `SELECT pp.amount, jsonb_build_object('product_id',pd.product_id,'product_name',p.name,'price_id',pp.id,'label',pp.label,
			'image_url',(SELECT '/media/images/products/' || pi.product_id::text || '/' || pi.filename FROM product_images pi WHERE pi.product_id=pd.product_id ORDER BY pi.created_at ASC, pi.id ASC LIMIT 1)), i.id
			FROM product_prices pp JOIN product_details pd ON pd.id=pp.product_detail_id JOIN products p ON p.id=pd.product_id
			JOIN product_variants v ON v.product_price_id=pp.id JOIN inventories i ON i.product_variant_id=v.id
			WHERE pp.id=$1 ORDER BY v.created_at, v.id, i.created_at, i.id LIMIT 1 FOR UPDATE OF i`, items[i].ProductPriceID).
			Scan(&price, &snapshot, &inventoryID)
		if errors.Is(err, pgx.ErrNoRows) {
			return model.ErrPriceNotFound
		}
		if err != nil {
			return fmt.Errorf("read order price: %w", err)
		}
		items[i].UnitPrice = fmt.Sprintf("%.2f", price)
		items[i].LineTotal = fmt.Sprintf("%.2f", price*float64(items[i].Quantity))
		items[i].ProductSnapshot = snapshot
		var itemID string
		err = tx.QueryRow(ctx, `INSERT INTO order_items(order_id,quantity,unit_price,line_total,product_snapshot) VALUES($1,$2,$3,$4,$5::jsonb) RETURNING id,created_at`, order.ID, items[i].Quantity, price, price*float64(items[i].Quantity), snapshot).Scan(&itemID, &items[i].CreatedAt)
		if err != nil {
			return fmt.Errorf("create order item: %w", err)
		}
		items[i].ID, items[i].OrderID = itemID, order.ID
		rows, qerr := tx.Query(ctx, `SELECT ii.id FROM inventory_items ii WHERE ii.inventory_id=$1 AND ii.status='可用' AND NOT EXISTS (SELECT 1 FROM inventory_reservations ir WHERE ir.inventory_item_id=ii.id AND ir.status='active') ORDER BY ii.created_at FOR UPDATE SKIP LOCKED LIMIT $2`, inventoryID, items[i].Quantity)
		if qerr != nil {
			return qerr
		}
		ids, qerr := pgx.CollectRows(rows, func(row pgx.CollectableRow) (string, error) { var id string; e := row.Scan(&id); return id, e })
		if qerr != nil {
			return qerr
		}
		if len(ids) != items[i].Quantity {
			return model.ErrInsufficientInventory
		}
		for _, itemID := range ids {
			if _, err = tx.Exec(ctx, `INSERT INTO inventory_reservations(order_id,order_item_id,inventory_item_id) VALUES($1,$2,$3)`, order.ID, items[i].ID, itemID); err != nil {
				return fmt.Errorf("reserve inventory: %w", err)
			}
		}
		subtotal += price * float64(items[i].Quantity)
	}
	if err = tx.QueryRow(ctx, `UPDATE orders SET subtotal=$2,total_amount=$2,updated_at=now() WHERE id=$1 RETURNING subtotal,total_amount,updated_at`, order.ID, subtotal).Scan(&order.Subtotal, &order.TotalAmount, &order.UpdatedAt); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO order_status_history(order_id,status) VALUES($1,'pending')`, order.ID); err != nil {
		return err
	}
	order.Items = items
	return tx.Commit(ctx)
}

func scanOrder(row pgx.Row) (*model.Order, error) {
	var o model.Order
	err := row.Scan(&o.ID, &o.OrderNo, &o.CustomerID, &o.Status, &o.PaymentStatus, &o.FulfillmentStatus, &o.Subtotal, &o.TotalAmount, &o.CustomerSnapshot, &o.ShippingAddressSnapshot, &o.CreatedAt, &o.UpdatedAt)
	return &o, err
}
func (r *OrderRepositoryPGX) get(ctx context.Context, query string, args ...any) (*model.Order, error) {
	o, err := scanOrder(r.pool.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	o.Items, err = r.items(ctx, o.ID)
	return o, err
}
func (r *OrderRepositoryPGX) items(ctx context.Context, orderID string) ([]model.OrderItem, error) {
	rows, err := r.pool.Query(ctx, `SELECT id,order_id,quantity,unit_price,line_total,product_snapshot,created_at FROM order_items WHERE order_id=$1 ORDER BY created_at`, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.OrderItem, error) {
		var i model.OrderItem
		err := row.Scan(&i.ID, &i.OrderID, &i.Quantity, &i.UnitPrice, &i.LineTotal, &i.ProductSnapshot, &i.CreatedAt)
		return i, err
	})
}

const orderSelect = `SELECT id,order_no,customer_id,status,payment_status,fulfillment_status,subtotal,total_amount,customer_snapshot,shipping_address_snapshot,created_at,updated_at FROM orders`

func (r *OrderRepositoryPGX) GetByID(ctx context.Context, id, memberID string, employee bool) (*model.Order, error) {
	q := orderSelect + ` WHERE id=$1`
	args := []any{id}
	if !employee {
		q += ` AND customer_id=$2`
		args = append(args, memberID)
	}
	return r.get(ctx, q, args...)
}
func (r *OrderRepositoryPGX) List(ctx context.Context, memberID, status string, page, size int, employee bool) ([]model.Order, int, error) {
	where := " WHERE 1=1"
	args := []any{}
	n := 1
	if !employee {
		where += fmt.Sprintf(" AND customer_id=$%d", n)
		args = append(args, memberID)
		n++
	}
	if status != "" {
		where += fmt.Sprintf(" AND status=$%d", n)
		args = append(args, status)
		n++
	}
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT count(*) FROM orders"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	q := orderSelect + where + fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", n, n+1)
	args = append(args, size, (page-1)*size)
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	orders, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.Order, error) {
		o := model.Order{}
		e := row.Scan(&o.ID, &o.OrderNo, &o.CustomerID, &o.Status, &o.PaymentStatus, &o.FulfillmentStatus, &o.Subtotal, &o.TotalAmount, &o.CustomerSnapshot, &o.ShippingAddressSnapshot, &o.CreatedAt, &o.UpdatedAt)
		if e == nil {
			o.Items, e = r.items(ctx, o.ID)
		}
		return o, e
	})
	return orders, total, err
}
func (r *OrderRepositoryPGX) Cancel(ctx context.Context, id, memberID string, employee bool) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := `UPDATE orders SET status='cancelled',fulfillment_status='cancelled',updated_at=now() WHERE id=$1 AND status='pending'`
	args := []any{id}
	if !employee {
		q += ` AND customer_id=$2`
		args = append(args, memberID)
	}
	tag, err := tx.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return model.ErrOrderNotCancellable
	}
	if _, err = tx.Exec(ctx, `UPDATE inventory_reservations SET status='released',released_at=now() WHERE order_id=$1 AND status='active'`, id); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO order_status_history(order_id,status,changed_by) VALUES($1,'cancelled',$2)`, id, memberID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (r *OrderRepositoryPGX) UpdateStatus(ctx context.Context, id, memberID, status string) (*model.Order, error) {
	if status != "confirmed" && status != "completed" {
		return nil, model.ErrInvalidOrderTransition
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var old string
	err = tx.QueryRow(ctx, `SELECT status FROM orders WHERE id=$1 FOR UPDATE`, id).Scan(&old)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, model.ErrOrderNotFound
	}
	if err != nil {
		return nil, err
	}
	if (old != "pending" || status != "confirmed") && (old != "confirmed" || status != "completed") {
		return nil, model.ErrInvalidOrderTransition
	}
	if _, err = tx.Exec(ctx, `UPDATE orders SET status=$2,updated_at=now() WHERE id=$1`, id, status); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO order_status_history(order_id,status,changed_by) VALUES($1,$2,$3)`, id, status, memberID); err != nil {
		return nil, err
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id, memberID, true)
}
func (r *OrderRepositoryPGX) History(ctx context.Context, id, memberID string, employee bool) ([]model.OrderStatusHistory, error) {
	if _, err := r.GetByID(ctx, id, memberID, employee); err != nil {
		return nil, err
	}
	rows, err := r.pool.Query(ctx, `SELECT id,order_id,status,changed_by,created_at FROM order_status_history WHERE order_id=$1 ORDER BY created_at`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (model.OrderStatusHistory, error) {
		var h model.OrderStatusHistory
		var changedBy *string
		err := row.Scan(&h.ID, &h.OrderID, &h.Status, &changedBy, &h.CreatedAt)
		if changedBy != nil {
			h.ChangedBy = *changedBy
		}
		return h, err
	})
}
