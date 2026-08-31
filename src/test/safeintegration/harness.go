// Package safeintegration provides isolated PostgreSQL integration-test schemas.
package safeintegration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Harness struct {
	Pool      *pgxpool.Pool
	adminPool *pgxpool.Pool
	schema    string
	sentinel  string
}

const isolatedRootID = "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

func Open(ctx context.Context, databaseURL string) (*Harness, error) {
	u, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	adminConfig := *u
	adminConfig.ConnConfig.Database = "productdb"
	admin, err := pgxpool.NewWithConfig(ctx, &adminConfig)
	if err != nil {
		return nil, fmt.Errorf("connect to productdb: %w", err)
	}
	var sentinel string
	if err := admin.QueryRow(ctx, `SELECT row_to_json(m)::text FROM public.members m WHERE email = 'root@gmail.com'`).Scan(&sentinel); err != nil {
		admin.Close()
		return nil, fmt.Errorf("read public sentinel: %w", err)
	}

	schema := "integration_" + uuid.NewString()
	if _, err = admin.Exec(ctx, "CREATE SCHEMA "+pgx.Identifier{schema}.Sanitize()); err != nil {
		admin.Close()
		return nil, fmt.Errorf("create isolated schema: %w", err)
	}

	u.ConnConfig.RuntimeParams["search_path"] = pgx.Identifier{schema}.Sanitize()
	u.MaxConns = 4
	u.MinConns = 1
	u.MaxConnLifetime = 5 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, u)
	if err != nil {
		_, _ = admin.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{schema}.Sanitize()+" CASCADE")
		admin.Close()
		return nil, fmt.Errorf("connect with isolated search path: %w", err)
	}
	h := &Harness{Pool: pool, adminPool: admin, schema: schema, sentinel: sentinel}
	if err := h.migrate(ctx); err != nil {
		h.Close(ctx)
		return nil, err
	}
	return h, nil
}

func (h *Harness) migrate(ctx context.Context) error {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../../"))
	entries, err := os.ReadDir(filepath.Join(root, "db/migrations"))
	if err != nil {
		return err
	}
	for _, entry := range entries {
		// Migration 014 seeds the public administrator. The isolated schema must
		// not create a second fixture member; the public row is the sentinel.
		if entry.Name() == "014_insert_admin_member.up.sql" {
			continue
		}
		if filepath.Ext(entry.Name()) != ".sql" || filepath.Base(entry.Name())[0:3] == "000" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(root, "db/migrations", entry.Name()))
		if err != nil {
			return err
		}
		if _, err = h.Pool.Exec(ctx, string(data)); err != nil {
			return fmt.Errorf("migration %s: %w", entry.Name(), err)
		}
	}
	if _, err := h.Pool.Exec(ctx, `
		INSERT INTO members (id, email, password, name, member_type, permission)
		VALUES ($1, 'root@gmail.com', '$2a$10$8cvP4Nv3LdR3J303AQ7NIOSnb1rQaNU/iyo65Gcv/oFSTyP03UodK', 'root', 'employee', 'admin')`, isolatedRootID); err != nil {
		return fmt.Errorf("seed isolated root member: %w", err)
	}
	return nil
}

func (h *Harness) Reset(ctx context.Context) error {
	quoted := pgx.Identifier{h.schema}.Sanitize()
	if _, err := h.adminPool.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
		return err
	}
	if _, err := h.adminPool.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		return err
	}
	return h.migrate(ctx)
}

func (h *Harness) Close(ctx context.Context) error {
	if h == nil {
		return nil
	}
	if h.Pool != nil {
		h.Pool.Close()
	}
	var err error
	if h.adminPool != nil {
		_, err = h.adminPool.Exec(ctx, "DROP SCHEMA "+pgx.Identifier{h.schema}.Sanitize()+" CASCADE")
		var sentinel string
		if scanErr := h.adminPool.QueryRow(ctx, `SELECT row_to_json(m)::text FROM public.members m WHERE email = 'root@gmail.com'`).Scan(&sentinel); scanErr != nil {
			if err == nil {
				err = scanErr
			}
		} else if sentinel != h.sentinel {
			if err == nil {
				err = fmt.Errorf("public sentinel changed during integration tests")
			}
		}
		h.adminPool.Close()
	}
	return err
}
