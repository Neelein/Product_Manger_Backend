package main

import (
	"context"
	"log"
	"os"
	"time"

	"backend/src/test/storefrontseed"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = storefrontseed.DefaultDatabaseURL
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to storefront E2E database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping storefront E2E database: %v", err)
	}
	if err := storefrontseed.Seed(ctx, pool); err != nil {
		log.Fatalf("seed storefront E2E data: %v", err)
	}
	if err := storefrontseed.Verify(ctx, pool); err != nil {
		log.Fatalf("verify storefront E2E data: %v", err)
	}

	log.Printf("storefront E2E fixture is ready in %s (product %s)", databaseURL, storefrontseed.ProductID)
}
