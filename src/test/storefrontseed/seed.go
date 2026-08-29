package storefrontseed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultDatabaseURL = "postgres://root:root123@localhost:5432/productdb_storefront_e2e?sslmode=disable"

const (
	ProductID  = "11111111-1111-1111-1111-111111111111"
	DetailID   = "22222222-2222-2222-2222-222222222222"
	PriceAID   = "33333333-3333-3333-3333-333333333333"
	PriceBID   = "44444444-4444-4444-4444-444444444444"
	CategoryID = "55555555-5555-5555-5555-555555555555"
)

// Seed upserts the deterministic data used by the storefront browser E2E.
// It intentionally does not remove any other data from the database.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin storefront seed transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	statements := []struct {
		query string
		args  []any
	}{
		{
			`INSERT INTO categories (id, name) VALUES ($1, $2)
			 ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name`,
			[]any{CategoryID, "Integration Category"},
		},
		{
			`INSERT INTO products (id, type, name, status, category_id, member_id)
			 VALUES ($1, 'product', $2, 'active', $3, '00000000-0000-0000-0000-000000000000')
			 ON CONFLICT (id) DO UPDATE SET
			 name = EXCLUDED.name, status = EXCLUDED.status, category_id = EXCLUDED.category_id`,
			[]any{ProductID, "Integration Product", CategoryID},
		},
		{
			`INSERT INTO product_details (id, product_id, introduction, usage_instructions, return_policy)
			 VALUES ($1, $2, $3, $4, $5)
			 ON CONFLICT (id) DO UPDATE SET
			 product_id = EXCLUDED.product_id, introduction = EXCLUDED.introduction,
			 usage_instructions = EXCLUDED.usage_instructions, return_policy = EXCLUDED.return_policy`,
			[]any{DetailID, ProductID, "Integration introduction", "Integration usage", "Integration returns"},
		},
		{
			`INSERT INTO product_prices (id, product_detail_id, label, amount, currency, sort_order)
			 VALUES ($1, $2, $3, $4, 'TWD', $5)
			 ON CONFLICT (id) DO UPDATE SET
			 product_detail_id = EXCLUDED.product_detail_id, label = EXCLUDED.label,
			 amount = EXCLUDED.amount, currency = EXCLUDED.currency, sort_order = EXCLUDED.sort_order`,
			[]any{PriceAID, DetailID, "Standard", 125.50, 1},
		},
		{
			`INSERT INTO product_prices (id, product_detail_id, label, amount, currency, sort_order)
			 VALUES ($1, $2, $3, $4, 'TWD', $5)
			 ON CONFLICT (id) DO UPDATE SET
			 product_detail_id = EXCLUDED.product_detail_id, label = EXCLUDED.label,
			 amount = EXCLUDED.amount, currency = EXCLUDED.currency, sort_order = EXCLUDED.sort_order`,
			[]any{PriceBID, DetailID, "Premium", 250, 2},
		},
	}

	for _, statement := range statements {
		if _, err := tx.Exec(ctx, statement.query, statement.args...); err != nil {
			return fmt.Errorf("execute storefront seed statement: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit storefront seed transaction: %w", err)
	}
	return nil
}

// Verify confirms that the public product fixture has the expected rows and values.
func Verify(ctx context.Context, pool *pgxpool.Pool) error {
	var categoryName, productName, productStatus, productCategoryID string
	var detailIntroduction, detailUsage, detailReturns string
	var priceCount int
	if err := pool.QueryRow(ctx, `
		SELECT c.name, p.name, p.status, p.category_id::text,
		       pd.introduction, pd.usage_instructions, pd.return_policy,
		       (SELECT COUNT(*) FROM product_prices WHERE product_detail_id = pd.id)
		FROM products p
		JOIN categories c ON c.id = p.category_id
		JOIN product_details pd ON pd.product_id = p.id
		WHERE p.id = $1 AND pd.id = $2`, ProductID, DetailID).Scan(
		&categoryName, &productName, &productStatus, &productCategoryID,
		&detailIntroduction, &detailUsage, &detailReturns, &priceCount,
	); err != nil {
		return fmt.Errorf("read seeded storefront product: %w", err)
	}

	if categoryName != "Integration Category" || productName != "Integration Product" ||
		productStatus != "active" || productCategoryID != CategoryID ||
		detailIntroduction != "Integration introduction" || detailUsage != "Integration usage" ||
		detailReturns != "Integration returns" || priceCount != 2 {
		return fmt.Errorf("seeded storefront product does not match expected fixture")
	}
	return nil
}
