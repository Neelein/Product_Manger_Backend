package database

import (
	"context"
	"errors"
	"fmt"

	"backend/src/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MemberRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewMemberRepositoryPGX(pool *pgxpool.Pool) *MemberRepositoryPGX {
	return &MemberRepositoryPGX{pool: pool}
}

func (r *MemberRepositoryPGX) Create(ctx context.Context, member *domain.Member) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO members (email, password, name)
		 VALUES ($1, $2, $3)
		 RETURNING id, created_at, updated_at`,
		member.Email, member.Password, member.Name,
	).Scan(&member.ID, &member.CreatedAt, &member.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("creating member: %w", err)
	}
	return nil
}

func (r *MemberRepositoryPGX) GetByEmail(ctx context.Context, email string) (*domain.Member, error) {
	var m domain.Member
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, name, created_at, updated_at
		 FROM members WHERE email = $1`, email,
	).Scan(&m.ID, &m.Email, &m.Password, &m.Name, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting member by email: %w", err)
	}
	return &m, nil
}

func (r *MemberRepositoryPGX) GetByID(ctx context.Context, id string) (*domain.Member, error) {
	var m domain.Member
	err := r.pool.QueryRow(ctx,
		`SELECT id, email, password, name, created_at, updated_at
		 FROM members WHERE id = $1`, id,
	).Scan(&m.ID, &m.Email, &m.Password, &m.Name, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting member by ID: %w", err)
	}
	return &m, nil
}

func (r *MemberRepositoryPGX) Update(ctx context.Context, member *domain.Member) error {
	ct, err := r.pool.Exec(ctx,
		`UPDATE members
		 SET email = $1, name = $2, updated_at = now()
		 WHERE id = $3`,
		member.Email, member.Name, member.ID,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("updating member: %w", err)
	}

	if ct.RowsAffected() == 0 {
		return domain.ErrMemberNotFound
	}

	err = r.pool.QueryRow(ctx,
		`SELECT updated_at FROM members WHERE id = $1`, member.ID,
	).Scan(&member.UpdatedAt)
	if err != nil {
		return fmt.Errorf("reading updated timestamp: %w", err)
	}

	return nil
}
