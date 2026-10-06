package main

import (
	"errors"
	"testing"
)

func TestRunMigrationCommandDoesNotRequirePaymentOTP(t *testing.T) {
	const wantDatabaseURL = "postgres://localhost:5432/productdb?sslmode=disable"
	called := false

	err := runMigrationCommand(func(key string) string {
		switch key {
		case "DATABASE_URL":
			return wantDatabaseURL
		case "PAYMENT_OTP":
			t.Fatal("migration command must not read PAYMENT_OTP")
			return ""
		default:
			return ""
		}
	}, func(databaseURL string) error {
		called = true
		if databaseURL != wantDatabaseURL {
			t.Fatalf("migration command database URL = %q, want %q", databaseURL, wantDatabaseURL)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("runMigrationCommand() error = %v", err)
	}
	if !called {
		t.Fatal("migration runner was not called")
	}
}

func TestRunMigrationCommandRequiresDatabaseURL(t *testing.T) {
	want := "DATABASE_URL is not set"
	err := runMigrationCommand(func(string) string { return "  " }, func(string) error {
		return errors.New("migration runner should not be called")
	})
	if err == nil || err.Error() != want {
		t.Fatalf("runMigrationCommand() error = %v, want %q", err, want)
	}
}
