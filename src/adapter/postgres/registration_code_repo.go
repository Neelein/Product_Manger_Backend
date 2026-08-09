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

type RegistrationCodeRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewRegistrationCodeRepositoryPGX(pool *pgxpool.Pool) *RegistrationCodeRepositoryPGX {
	return &RegistrationCodeRepositoryPGX{pool: pool}
}

func (r *RegistrationCodeRepositoryPGX) RegisterMemberWithCode(ctx context.Context, member *domain.Member, code string) error {
	err := r.pool.QueryRow(ctx, "SELECT * FROM register_member_with_code($1, $2, $3, $4)",
		member.Email, member.Password, member.Name, code,
	).Scan(&member.ID, &member.Role, &member.CreatedAt, &member.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return domain.ErrEmailAlreadyExists
			case "R0001":
				return domain.ErrInvalidRegistrationCode
			case "R0002":
				return domain.ErrRegistrationCodeUsed
			}
		}
		return fmt.Errorf("registering member with code: %w", err)
	}
	return nil
}

func (r *RegistrationCodeRepositoryPGX) Create(ctx context.Context, createdBy string, code string) (*domain.RegistrationCode, error) {
	rc := &domain.RegistrationCode{Code: code}
	var cb *string
	if createdBy != "" {
		cb = &createdBy
	}

	err := r.pool.QueryRow(ctx, "SELECT * FROM create_registration_code($1, $2)", cb, code).Scan(&rc.ID, &rc.Code, &rc.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("creating registration code: duplicate code: %w", err)
		}
		return nil, fmt.Errorf("creating registration code: %w", err)
	}
	rc.CreatedBy = createdBy
	rc.Status = "available"
	return rc, nil
}

func (r *RegistrationCodeRepositoryPGX) List(ctx context.Context) ([]domain.RegistrationCode, error) {
	rows, err := r.pool.Query(ctx, "SELECT * FROM list_registration_codes()")
	if err != nil {
		return nil, fmt.Errorf("listing registration codes: %w", err)
	}
	defer rows.Close()

	codes := []domain.RegistrationCode{}
	for rows.Next() {
		var rc domain.RegistrationCode
		var usedAt *time.Time
		if err := rows.Scan(
			&rc.ID,
			&rc.Code,
			&rc.CreatedBy,
			&rc.CreatedByEmail,
			&rc.UsedBy,
			&rc.UsedByEmail,
			&usedAt,
			&rc.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning registration code: %w", err)
		}
		rc.UsedAt = usedAt
		rc.Status = "available"
		if usedAt != nil {
			rc.Status = "used"
		}
		codes = append(codes, rc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating registration codes: %w", err)
	}
	return codes, nil
}

func (r *RegistrationCodeRepositoryPGX) Delete(ctx context.Context, id string) (bool, error) {
	var deleted bool
	err := r.pool.QueryRow(ctx, "SELECT * FROM delete_registration_code($1)", id).Scan(&deleted)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("deleting registration code: %w", err)
	}
	return deleted, nil
}
