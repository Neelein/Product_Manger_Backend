package main

import (
	"context"
	"log"
	"os"
	"time"

	"backend/src/test/integrationseed"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL must target the disposable productdb")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatalf("connect to integration productdb: %v", err)
	}
	defer pool.Close()
	if err := integrationseed.Seed(ctx, pool); err != nil {
		log.Fatal(err)
	}
	if err := integrationseed.Verify(ctx, pool); err != nil {
		log.Fatal(err)
	}
}
