package database

import (
	"context"
	"errors"
	"fmt"

	"backend/src/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProductRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewProductRepositoryPGX(pool *pgxpool.Pool) *ProductRepositoryPGX {
	return &ProductRepositoryPGX{pool: pool}
}

func (r *ProductRepositoryPGX) Create(
	ctx context.Context,
	product *domain.Product,
) error {
	status := product.Status
	if status == "" {
		status = "active"
	}

	memberID := product.CreatedBy
	if memberID == "" {
		memberID = "00000000-0000-0000-0000-000000000000"
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product($1, $2, $3, $4)",
		product.Name, status, product.Category, memberID,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating product: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) List(
	ctx context.Context,
) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_products()")
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}

	products, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Product, error) {
		var p domain.Product

		err := row.Scan(
			&p.ID, &p.Name, &p.Status, &p.Category,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return p, err
		}

		return p, nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}

	if products == nil {
		products = []domain.Product{}
	}

	return products, nil
}

func (r *ProductRepositoryPGX) GetByID(
	ctx context.Context,
	id string,
) (*domain.Product, error) {
	var p domain.Product

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_by_id($1)", id,
	).Scan(&p.ID, &p.Name, &p.Status, &p.Category, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("getting product by ID: %w", err)
	}

	return &p, nil
}

func (r *ProductRepositoryPGX) Update(
	ctx context.Context,
	product *domain.Product,
) error {
	status := product.Status
	if status == "" {
		status = "active"
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product($1, $2, $3, $4)",
		product.ID, product.Name, status, product.Category,
	).Scan(&product.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrProductNotFound
		}
		return fmt.Errorf("updating product: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) Delete(
	ctx context.Context,
	id string,
) error {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_product($1)", id).Scan(&deleted)
	if err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}
	if !deleted {
		return domain.ErrProductNotFound
	}

	return nil
}

func (r *ProductRepositoryPGX) CreateDetail(
	ctx context.Context,
	detail *domain.ProductDetail,
) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_detail($1, $2, $3, $4)",
		detail.ProductID, detail.Introduction, detail.UsageInstructions, detail.ReturnPolicy,
	).Scan(&detail.ID, &detail.CreatedAt, &detail.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating detail: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) GetDetailByProductID(
	ctx context.Context,
	productID string,
) (*domain.ProductDetail, error) {
	var d domain.ProductDetail

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_detail_by_product($1)", productID,
	).Scan(&d.ID, &d.ProductID, &d.Introduction, &d.UsageInstructions,
		&d.ReturnPolicy, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDetailNotFound
		}
		return nil, fmt.Errorf("getting detail by product ID: %w", err)
	}

	return &d, nil
}

func (r *ProductRepositoryPGX) UpdateDetail(
	ctx context.Context,
	detail *domain.ProductDetail,
) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product_detail($1, $2, $3, $4)",
		detail.ID, detail.Introduction, detail.UsageInstructions, detail.ReturnPolicy,
	).Scan(&detail.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrDetailNotFound
		}
		return fmt.Errorf("updating detail: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) GetPriceByID(
	ctx context.Context,
	id string,
) (*domain.ProductPrice, error) {
	var p domain.ProductPrice

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_price_by_id($1)", id,
	).Scan(&p.ID, &p.ProductDetailID, &p.Label, &p.Amount, &p.Currency,
		&p.SortOrder, &p.InventoryID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrPriceNotFound
		}
		return nil, fmt.Errorf("getting price by ID: %w", err)
	}

	return &p, nil
}

func (r *ProductRepositoryPGX) GetPricesByDetailID(
	ctx context.Context,
	detailID string,
) ([]domain.ProductPrice, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_product_prices_by_detail($1)", detailID)
	if err != nil {
		return nil, fmt.Errorf("listing prices by detail ID: %w", err)
	}

	prices, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.ProductPrice, error) {
		var p domain.ProductPrice
		err := row.Scan(&p.ID, &p.ProductDetailID, &p.Label, &p.Amount,
			&p.Currency, &p.SortOrder, &p.InventoryID, &p.CreatedAt, &p.UpdatedAt)
		return p, err
	})
	if err != nil {
		return nil, fmt.Errorf("listing prices by detail ID: %w", err)
	}

	if prices == nil {
		prices = []domain.ProductPrice{}
	}

	return prices, nil
}

func (r *ProductRepositoryPGX) UpdatePrice(
	ctx context.Context,
	price *domain.ProductPrice,
) error {
	var currency *string
	if price.Currency == "" {
		currency = nil
	} else {
		currency = &price.Currency
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product_price($1, $2, $3, $4, $5)",
		price.ID, price.Label, price.Amount, currency, price.SortOrder,
	).Scan(&price.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrPriceNotFound
		}
		return fmt.Errorf("updating price: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) CreatePrice(
	ctx context.Context,
	price *domain.ProductPrice,
) error {
	var currency *string
	if price.Currency == "" {
		currency = nil
	} else {
		currency = &price.Currency
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_price($1, $2, $3, $4, $5)",
		price.ProductDetailID, price.Label, price.Amount, currency, price.SortOrder,
	).Scan(&price.ID, &price.CreatedAt, &price.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating price: %w", err)
	}

	return nil
}
