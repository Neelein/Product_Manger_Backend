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
	var description, category *string
	if product.Description != "" {
		description = &product.Description
	}
	if product.Category != "" {
		category = &product.Category
	}

	var memberID *string
	if product.CreatedBy != "" {
		memberID = &product.CreatedBy
	}

	err := r.pool.QueryRow(ctx,
		`INSERT INTO products (type, name, description, category, member_id)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		"product", product.Name, description, category, memberID,
	).Scan(&product.ID, &product.CreatedAt, &product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("creating product: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) List(
	ctx context.Context,
) ([]domain.Product, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, name, description, category, created_at, updated_at
		 FROM products
		 ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	defer rows.Close()

	var products []domain.Product
	for rows.Next() {
		var (
			p                 domain.Product
			description, cat *string
		)

		err := rows.Scan(
			&p.ID, &p.Name, &description, &cat,
			&p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning product row: %w", err)
		}

		if description != nil {
			p.Description = *description
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
		p                 domain.Product
		description, cat *string
	)

	err := r.pool.QueryRow(ctx,
		`SELECT id, name, description, category, created_at, updated_at
		 FROM products WHERE id = $1`, id,
	).Scan(&p.ID, &p.Name, &description, &cat, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProductNotFound
		}
		return nil, fmt.Errorf("getting product by ID: %w", err)
	}

	if description != nil {
		p.Description = *description
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
	var description, category *string
	if product.Description != "" {
		description = &product.Description
	}
	if product.Category != "" {
		category = &product.Category
	}

	ct, err := r.pool.Exec(ctx,
		`UPDATE products
		 SET name = $1, description = $2, category = $3, updated_at = now()
		 WHERE id = $4`,
		product.Name, description, category, product.ID,
	)
	if err != nil {
		return fmt.Errorf("updating product: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}

	err = r.pool.QueryRow(ctx,
		`SELECT updated_at FROM products WHERE id = $1`, product.ID,
	).Scan(&product.UpdatedAt)
	if err != nil {
		return fmt.Errorf("reading updated timestamp: %w", err)
	}

	return nil
}

func (r *ProductRepositoryPGX) Delete(
	ctx context.Context,
	id string,
) error {
	ct, err := r.pool.Exec(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("deleting product: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return domain.ErrProductNotFound
	}

	return nil
}
