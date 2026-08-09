package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	domain "backend/src/domain/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CategoryRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewCategoryRepositoryPGX(pool *pgxpool.Pool) *CategoryRepositoryPGX {
	return &CategoryRepositoryPGX{pool: pool}
}

func (r *CategoryRepositoryPGX) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_categories()")
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}

	categories, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (domain.Category, error) {
		var c domain.Category

		err := row.Scan(
			&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return c, err
		}

		return c, nil
	})
	if err != nil {
		return nil, fmt.Errorf("listing categories: %w", err)
	}

	if categories == nil {
		categories = []domain.Category{}
	}

	return categories, nil
}

func (r *CategoryRepositoryPGX) Create(ctx context.Context, name string) (*domain.Category, error) {
	c := &domain.Category{Name: name}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_category($1)", name).Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, domain.ErrCategoryNameExists
		}
		return nil, fmt.Errorf("creating category: %w", err)
	}

	return c, nil
}

func (r *CategoryRepositoryPGX) Update(ctx context.Context, id, name string) (bool, error) {
	var updatedAt time.Time

	err := r.pool.QueryRow(ctx, "SELECT * FROM update_category($1, $2)", id, name).Scan(&updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return false, domain.ErrCategoryNameExists
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("updating category: %w", err)
	}

	return true, nil
}

func (r *CategoryRepositoryPGX) Delete(ctx context.Context, id string) (bool, error) {
	var deleted bool

	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_category($1)", id).Scan(&deleted)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "R0010" {
			return false, domain.ErrCategoryInUse
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("deleting category: %w", err)
	}

	return deleted, nil
}
