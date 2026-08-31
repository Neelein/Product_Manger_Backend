package config

import "testing"

func TestRequiredTestDatabaseURLRequiresExplicitValue(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	_, err := RequiredTestDatabaseURL()
	if err == nil {
		t.Fatal("RequiredTestDatabaseURL() error = nil, want missing configuration error")
	}
	if err.Error() != "DATABASE_URL must be explicitly set to productdb for integration tests" {
		t.Fatalf("RequiredTestDatabaseURL() error = %q, want clear missing configuration error", err)
	}
}

func TestRequiredTestDatabaseURLAcceptsProductDB(t *testing.T) {
	want := "postgres://root:root123@localhost:5432/productdb?sslmode=disable"
	t.Setenv("DATABASE_URL", want)

	got, err := RequiredTestDatabaseURL()
	if err != nil || got != want {
		t.Fatalf("RequiredTestDatabaseURL() = %q, %v, want productdb URL", got, err)
	}
}

func TestRequiredTestDatabaseURLRequiresDatabaseName(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://root:root123@localhost:5432/?sslmode=disable")

	_, err := RequiredTestDatabaseURL()
	if err == nil {
		t.Fatal("RequiredTestDatabaseURL() error = nil, want database name validation error")
	}
	if err.Error() != "DATABASE_URL must include the productdb database name" {
		t.Fatalf("RequiredTestDatabaseURL() error = %q, want database name validation error", err)
	}
}

func TestRequiredTestDatabaseURLAcceptsDedicatedDatabase(t *testing.T) {
	want := "postgres://root:root123@localhost:5432/productdb?sslmode=disable"
	t.Setenv("DATABASE_URL", want)

	got, err := RequiredTestDatabaseURL()
	if err != nil {
		t.Fatalf("RequiredTestDatabaseURL() error = %v, want nil", err)
	}
	if got != want {
		t.Fatalf("RequiredTestDatabaseURL() = %q, want %q", got, want)
	}
}
