package database

import (
	"context"
	"errors"
	"fmt"

	"backend/src/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionRepositoryPGX struct {
	pool *pgxpool.Pool
}

func NewSessionRepositoryPGX(pool *pgxpool.Pool) *SessionRepositoryPGX {
	return &SessionRepositoryPGX{pool: pool}
}

func (r *SessionRepositoryPGX) Create(ctx context.Context, session *domain.Session) error {
	err := r.pool.QueryRow(ctx,
		`INSERT INTO sessions (member_id)
		 VALUES ($1)
		 RETURNING id, session_key, created_at, expires_at`,
		session.MemberID,
	).Scan(&session.ID, &session.SessionKey, &session.CreatedAt, &session.ExpiresAt)
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	return nil
}

func (r *SessionRepositoryPGX) GetByKey(ctx context.Context, sessionKey string) (*domain.Session, error) {
	var s domain.Session
	err := r.pool.QueryRow(ctx,
		`SELECT id, member_id, session_key, created_at, expires_at
		 FROM sessions WHERE session_key = $1`, sessionKey,
	).Scan(&s.ID, &s.MemberID, &s.SessionKey, &s.CreatedAt, &s.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("getting session by key: %w", err)
	}
	return &s, nil
}

func (r *SessionRepositoryPGX) Delete(ctx context.Context, sessionKey string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE session_key = $1`, sessionKey)
	if err != nil {
		return fmt.Errorf("deleting session: %w", err)
	}
	return nil
}

func (r *SessionRepositoryPGX) DeleteByMemberID(ctx context.Context, memberID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM sessions WHERE member_id = $1`, memberID)
	if err != nil {
		return fmt.Errorf("deleting sessions by member ID: %w", err)
	}
	return nil
}
