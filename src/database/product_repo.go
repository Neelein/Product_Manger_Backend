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
	var category *string
	if product.Category != "" {
		category = &product.Category
	}

	var memberID *string
	if product.CreatedBy != "" {
		memberID = &product.CreatedBy
	}

	status := product.Status
	if status == "" {
		status = "active"
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product($1, $2, $3, $4)",
		product.Name, status, category, memberID,
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
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var (
			p   domain.Product
			cat *string
		)

		err := rows.Scan(
			&p.ID, &p.Name, &p.Status, &cat,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning product row: %w", err)
		}

		if cat != nil {
			p.Category = *cat
		}

		products = append(products, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating product rows: %w", err)
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
	var (
		p   domain.Product
		cat *string
	)

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_by_id($1)", id,
	).Scan(&p.ID, &p.Name, &p.Status, &cat, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("getting product by ID: %w", err)
	}

	if cat != nil {
		p.Category = *cat
	}

	return &p, nil
}

func (r *ProductRepositoryPGX) Update(
	ctx context.Context,
	product *domain.Product,
) error {
	var category *string
	if product.Category != "" {
		category = &product.Category
	}

	status := product.Status
	if status == "" {
		status = "active"
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product($1, $2, $3, $4)",
		product.ID, product.Name, status, category,
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
	var introduction, usageInstructions, returnPolicy *string
	if detail.Introduction != "" {
		introduction = &detail.Introduction
	}
	if detail.UsageInstructions != "" {
		usageInstructions = &detail.UsageInstructions
	}
	if detail.ReturnPolicy != "" {
		returnPolicy = &detail.ReturnPolicy
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_product_detail($1, $2, $3, $4)",
		detail.ProductID, introduction, usageInstructions, returnPolicy,
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
	var (
		d                                 domain.ProductDetail
		introduction, usage, returnPolicy *string
	)

	err := r.pool.QueryRow(ctx, "SELECT * FROM get_product_detail_by_product($1)", productID,
	).Scan(&d.ID, &d.ProductID, &introduction, &usage,
		&returnPolicy, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrDetailNotFound
		}
		return nil, fmt.Errorf("getting detail by product ID: %w", err)
	}

	if introduction != nil {
		d.Introduction = *introduction
	}
	if usage != nil {
		d.UsageInstructions = *usage
	}
	if returnPolicy != nil {
		d.ReturnPolicy = *returnPolicy
	}

	return &d, nil
}

func (r *ProductRepositoryPGX) UpdateDetail(
	ctx context.Context,
	detail *domain.ProductDetail,
) error {
	var introduction, usageInstructions, returnPolicy *string
	if detail.Introduction != "" {
		introduction = &detail.Introduction
	}
	if detail.UsageInstructions != "" {
		usageInstructions = &detail.UsageInstructions
	}
	if detail.ReturnPolicy != "" {
		returnPolicy = &detail.ReturnPolicy
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_product_detail($1, $2, $3, $4)",
		detail.ID, introduction, usageInstructions, returnPolicy,
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
		&p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
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
	defer rows.Close()

	var prices []domain.ProductPrice
	for rows.Next() {
		var p domain.ProductPrice
		err := rows.Scan(&p.ID, &p.ProductDetailID, &p.Label, &p.Amount,
			&p.Currency, &p.SortOrder, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("scanning price row: %w", err)
		}
		prices = append(prices, p)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating price rows: %w", err)
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
