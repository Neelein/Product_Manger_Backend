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
	err := r.pool.QueryRow(ctx, "SELECT * FROM create_member($1, $2, $3)",
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
	err := r.pool.QueryRow(ctx, "SELECT * FROM get_member_by_email($1)", email,
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
	err := r.pool.QueryRow(ctx, "SELECT * FROM get_member_by_id($1)", id,
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
	err := r.pool.QueryRow(ctx, "SELECT * FROM update_member($1, $2, $3)",
		member.ID, member.Email, member.Name,
	).Scan(&member.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrMemberNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrEmailAlreadyExists
		}
		return fmt.Errorf("updating member: %w", err)
	}
	return nil
}
