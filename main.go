package main

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/src/api"
	"backend/src/database"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/gorilla/mux"
)

//go:embed db/migrations/*.up.sql
var migrationsFS embed.FS

type Response struct {
	Message string `json:"message"`
}

func runMigrations(databaseURL string) error {
	d, err := iofs.New(migrationsFS, "db/migrations")
	if err != nil {
		return fmt.Errorf("init migration source: %w", err)
	}
	pgx5URL := strings.Replace(databaseURL, "postgres://", "pgx5://", 1)
	m, err := migrate.NewWithSourceInstance("iofs", d, pgx5URL)
	if err != nil {
		return fmt.Errorf("init migration: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("run migration: %w", err)
	}
	return nil
}

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := database.NewPool(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Running database migrations...")
	if err := runMigrations(databaseURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	repo := database.NewProductRepositoryPGX(pool)
	inventoryRepo := database.NewInventoryRepositoryPGX(pool)
	memberRepo := database.NewMemberRepositoryPGX(pool)
	sessionRepo := database.NewSessionCache(24 * time.Hour)
	defer sessionRepo.Stop()

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/health", healthHandler).Methods("GET")

	api.RegisterProductRoutes(r, repo, memberRepo, sessionRepo)
	api.RegisterInventoryRoutes(r, inventoryRepo, memberRepo, sessionRepo)
	api.RegisterMemberRoutes(r, memberRepo, sessionRepo)
	api.RegisterAnnouncementRoutes(r, database.NewAnnouncementRepositoryPGX(pool), memberRepo, sessionRepo)
	api.RegisterChatRoutes(r, database.NewChatRoomRepositoryPGX(pool), memberRepo, sessionRepo)

	log.Println("Server starting on :8090")
	log.Fatal(http.ListenAndServe(":8090", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "Welcome to Product Manager API"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "OK"})
}
