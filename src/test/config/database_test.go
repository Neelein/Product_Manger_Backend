package config

import "testing"

func TestRequiredTestDatabaseURLRequiresExplicitValue(t *testing.T) {
	t.Setenv("TEST_DATABASE_URL", "")

	_, err := RequiredTestDatabaseURL()
	if err == nil {
		t.Fatal("RequiredTestDatabaseURL() error = nil, want missing configuration error")
	}
	if err.Error() != "TEST_DATABASE_URL must be explicitly set to a dedicated test database" {
		t.Fatalf("RequiredTestDatabaseURL() error = %q, want clear missing configuration error", err)
	}
}

func TestRequiredTestDatabaseURLRejectsProductionDatabase(t *testing.T) {
	t.Setenv("TEST_DATABASE_URL", "postgres://root:root123@localhost:5432/productdb?sslmode=disable")

	_, err := RequiredTestDatabaseURL()
	if err == nil {
		t.Fatal("RequiredTestDatabaseURL() error = nil, want production database rejection")
	}
	if err.Error() != "TEST_DATABASE_URL must not target the production productdb database" {
		t.Fatalf("RequiredTestDatabaseURL() error = %q, want production database rejection", err)
	}
}

func TestRequiredTestDatabaseURLRequiresDatabaseName(t *testing.T) {
	t.Setenv("TEST_DATABASE_URL", "postgres://root:root123@localhost:5432/?sslmode=disable")

	_, err := RequiredTestDatabaseURL()
	if err == nil {
		t.Fatal("RequiredTestDatabaseURL() error = nil, want database name validation error")
	}
	if err.Error() != "TEST_DATABASE_URL must include a dedicated test database name" {
		t.Fatalf("RequiredTestDatabaseURL() error = %q, want database name validation error", err)
	}
}

func TestRequiredTestDatabaseURLAcceptsDedicatedDatabase(t *testing.T) {
	want := "postgres://root:root123@localhost:5432/productdb_ci_test?sslmode=disable"
	t.Setenv("TEST_DATABASE_URL", want)

	got, err := RequiredTestDatabaseURL()
	if err != nil {
		t.Fatalf("RequiredTestDatabaseURL() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("RequiredTestDatabaseURL() = %q, want %q", got, want)
	}
}
