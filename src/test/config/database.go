package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func RequiredTestDatabaseURL() (string, error) {
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return "", fmt.Errorf("DATABASE_URL must be explicitly set to productdb for integration tests")
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("DATABASE_URL is invalid: %w", err)
	}
	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if databaseName == "" {
		return "", fmt.Errorf("DATABASE_URL must include the productdb database name")
	}
	if databaseName == "productdb" {
		return databaseURL, nil
	}
	return "", fmt.Errorf("DATABASE_URL must target productdb; isolated schema cleanup is used")
}
