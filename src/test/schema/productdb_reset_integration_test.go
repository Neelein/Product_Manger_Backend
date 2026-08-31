//go:build integration

package schema_test

import (
	"context"
	"net/url"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestProductDBResetAndRootSeed(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Fatal("DATABASE_URL must explicitly identify localhost:5432/productdb")
	}
	parsed, err := url.Parse(databaseURL)
	if err != nil || parsed.Hostname() != "localhost" || parsed.Port() != "5432" || parsed.Path != "/productdb" {
		t.Fatalf("refusing non-target database URL")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	var database, schema string
	if err := pool.QueryRow(ctx, "SELECT current_database(), current_schema()").Scan(&database, &schema); err != nil {
		t.Fatal(err)
	}
	if database != "productdb" || schema != "public" {
		t.Fatalf("unexpected database identity: %s.%s", database, schema)
	}

	var version int
	var dirty bool
	if err := pool.QueryRow(ctx, "SELECT version, dirty FROM public.schema_migrations").Scan(&version, &dirty); err != nil {
		t.Fatal(err)
	}
	if version != 22 || dirty {
		t.Fatalf("unexpected migration state: version=%d dirty=%t", version, dirty)
	}

	var email, name, memberType, permission string
	var hashLength int
	var bcrypt bool
	if err := pool.QueryRow(ctx, `
		SELECT email, name, member_type, permission,
		       password LIKE '$2%', length(password)
		FROM public.members WHERE email = 'root@gmail.com'`).Scan(
		&email, &name, &memberType, &permission, &bcrypt, &hashLength,
	); err != nil {
		t.Fatal(err)
	}
	if email != "root@gmail.com" || name != "root" || memberType != "employee" || permission != "admin" || !bcrypt || hashLength != 60 {
		t.Fatalf("unexpected root member contract: email=%s name=%s member_type=%s permission=%s bcrypt=%t length=%d", email, name, memberType, permission, bcrypt, hashLength)
	}
}
