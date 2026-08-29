package config

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

func RequiredTestDatabaseURL() (string, error) {
	databaseURL := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if databaseURL == "" {
		return "", fmt.Errorf("TEST_DATABASE_URL must be explicitly set to a dedicated test database")
	}

	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", fmt.Errorf("TEST_DATABASE_URL is invalid: %w", err)
	}
	databaseName := strings.TrimPrefix(parsed.Path, "/")
	if databaseName == "" {
		return "", fmt.Errorf("TEST_DATABASE_URL must include a dedicated test database name")
	}
	if databaseName == "productdb" {
		return "", fmt.Errorf("TEST_DATABASE_URL must not target the production productdb database")
	}

	return databaseURL, nil
}
