package integrationseed

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

const RootEmail = "root@gmail.com"
const rootPasswordHash = "$2a$10$8cvP4Nv3LdR3J303AQ7NIOSnb1rQaNU/iyo65Gcv/oFSTyP03UodK"

// Seed creates only disposable integration fixtures in the explicitly
// provisioned productdb. Production migrations intentionally contain none.
func Seed(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		INSERT INTO members (email, password, name, member_type, permission)
		VALUES ($1, $2, 'root', 'employee', 'admin')
		ON CONFLICT (email) DO UPDATE SET
			password = EXCLUDED.password, name = EXCLUDED.name,
			member_type = EXCLUDED.member_type, permission = EXCLUDED.permission`,
		RootEmail, rootPasswordHash)
	if err != nil {
		return fmt.Errorf("seed integration root member: %w", err)
	}
	return nil
}

func Verify(ctx context.Context, pool *pgxpool.Pool) error {
	var email, name, memberType, permission, password string
	if err := pool.QueryRow(ctx, `SELECT email, password, name, member_type, permission FROM members WHERE email = $1`, RootEmail).
		Scan(&email, &password, &name, &memberType, &permission); err != nil {
		return fmt.Errorf("verify integration root member: %w", err)
	}
	if email != RootEmail || name != "root" || memberType != "employee" || permission != "admin" || len(password) != 60 {
		return fmt.Errorf("integration root member does not match fixture contract")
	}
	return nil
}
