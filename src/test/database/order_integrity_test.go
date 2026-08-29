//go:build integration

package database_test

import (
	"context"
	"testing"

	database "backend/src/adapter/postgres"
	"backend/src/domain/model"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestOrderDatabaseIntegrity(t *testing.T) {
	ctx := context.Background()
	var customerID, employeeID, firstOrderID, secondOrderID, itemID, inventoryID, inventoryItemID string

	require.NoError(t, testPool.QueryRow(ctx, `
		INSERT INTO members (email, password, name, member_type)
		VALUES ('order-customer-integrity@example.com', 'pw', 'Customer', 'customer')
		RETURNING id`).Scan(&customerID))
	require.NoError(t, testPool.QueryRow(ctx, `
		INSERT INTO members (email, password, name, member_type, permission)
		VALUES ('order-employee-integrity@example.com', 'pw', 'Employee', 'employee', 'admin')
		RETURNING id`).Scan(&employeeID))
	var codeID string
	require.NoError(t, testPool.QueryRow(ctx, `
		INSERT INTO registration_codes (code, used_by, used_at)
		VALUES ('ORDER-INTEGRITY-CODE', $1, now()) RETURNING id`, customerID).Scan(&codeID))
	_, err := testPool.Exec(ctx, "DELETE FROM members WHERE id = $1", customerID)
	require.Error(t, err, "deleting a member must not clear a consumed registration code")
	require.NoError(t, testPool.QueryRow(ctx, "DELETE FROM registration_codes WHERE id = $1 RETURNING id", codeID).Scan(&codeID))
	defer func() { _, _ = testPool.Exec(ctx, "DELETE FROM members WHERE id IN ($1, $2)", customerID, employeeID) }()

	insertOrder := func(orderNo, memberID string) (string, error) {
		var id string
		err := testPool.QueryRow(ctx, `
			INSERT INTO orders (order_no, customer_id, subtotal, total_amount, customer_snapshot, shipping_address_snapshot)
			VALUES ($1, $2, 10, 10, '{}'::jsonb, '{}'::jsonb) RETURNING id`, orderNo, memberID).Scan(&id)
		return id, err
	}
	firstOrderID, err = insertOrder("ORDER-INTEGRITY-1", customerID)
	require.NoError(t, err)
	_, err = insertOrder("ORDER-INTEGRITY-1", customerID)
	require.Error(t, err, "order numbers must be unique")
	_, err = insertOrder("ORDER-INTEGRITY-EMPLOYEE", employeeID)
	require.NoError(t, err, "employee members may create orders")
	secondOrderID, err = insertOrder("ORDER-INTEGRITY-2", customerID)
	require.NoError(t, err)

	require.NoError(t, testPool.QueryRow(ctx, `
		INSERT INTO order_items (order_id, quantity, unit_price, line_total, product_snapshot)
		VALUES ($1, 1, 10, 10, '{}'::jsonb) RETURNING id`, firstOrderID).Scan(&itemID))
	require.NoError(t, testPool.QueryRow(ctx, `
		INSERT INTO products (type, name) VALUES ('test', 'integrity product') RETURNING id`).Scan(new(string)))
	var detailID, priceID string
	var productID string
	require.NoError(t, testPool.QueryRow(ctx, "SELECT id FROM products WHERE name = 'integrity product' ORDER BY created_at DESC LIMIT 1").Scan(&productID))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO product_details (product_id) VALUES ($1) RETURNING id`, productID).Scan(&detailID))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO product_prices (product_detail_id, label, amount) VALUES ($1, 'default', 10) RETURNING id`, detailID).Scan(&priceID))
	var variantID string
	require.NoError(t, testPool.QueryRow(ctx, `SELECT id FROM create_product_variant($1, $2, NULL, 'active', NULL)`, detailID, priceID).Scan(&variantID))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO inventories (product_variant_id, product_price_id) VALUES ($1, $2) RETURNING id`, variantID, priceID).Scan(&inventoryID))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO inventory_items (inventory_id, item_code) VALUES ($1, 'integrity-item') RETURNING id`, inventoryID).Scan(&inventoryItemID))

	_, err = testPool.Exec(ctx, `
		INSERT INTO inventory_reservations (order_id, order_item_id, inventory_item_id)
		VALUES ($1, $2, $3)`, secondOrderID, itemID, inventoryItemID)
	require.Error(t, err, "reservation order and item order must match")
}

func TestOrderCreateRollsBackOrderAndItemsWhenInventoryReservationFails(t *testing.T) {
	ctx := context.Background()
	memberID, productID, detailID, priceID, variantID, inventoryID := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	defer func() {
		_, _ = testPool.Exec(ctx, "DELETE FROM products WHERE id=$1", productID)
		_, _ = testPool.Exec(ctx, "DELETE FROM members WHERE id=$1", memberID)
	}()
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO members(id,email,password,name,member_type) VALUES($1,$2,'pw','rollback','customer') RETURNING id`, memberID, memberID+"@example.com").Scan(new(string)))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO products(id,type,name) VALUES($1,'test','rollback product') RETURNING id`, productID).Scan(new(string)))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO product_details(id,product_id) VALUES($1,$2) RETURNING id`, detailID, productID).Scan(new(string)))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO product_prices(id,product_detail_id,label,amount) VALUES($1,$2,'default',12) RETURNING id`, priceID, detailID).Scan(new(string)))
	require.NoError(t, testPool.QueryRow(ctx, `SELECT id FROM create_product_variant($1,$2,NULL,'active',NULL)`, detailID, priceID).Scan(&variantID))
	require.NoError(t, testPool.QueryRow(ctx, `INSERT INTO inventories(id,product_variant_id,product_price_id) VALUES($1,$2,$3) RETURNING id`, inventoryID, variantID, priceID).Scan(new(string)))
	_, err := testPool.Exec(ctx, `INSERT INTO inventory_items(inventory_id,item_code) VALUES($1,'rollback-item')`, inventoryID)
	require.NoError(t, err)
	repo := database.NewOrderRepository(testPool)
	err = repo.Create(ctx, &model.Order{CustomerID: memberID, CustomerSnapshot: []byte(`{}`), ShippingAddressSnapshot: []byte(`{}`)}, []model.OrderItem{{ProductPriceID: priceID, Quantity: 2}})
	require.ErrorIs(t, err, model.ErrInsufficientInventory)
	var count int
	require.NoError(t, testPool.QueryRow(ctx, "SELECT count(*) FROM orders WHERE customer_id=$1", memberID).Scan(&count))
	require.Zero(t, count)
}
