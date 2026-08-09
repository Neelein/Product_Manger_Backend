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

	apphttp "backend/src/adapter/http"
	"backend/src/adapter/postgres"
	"backend/src/adapter/session"
	"backend/src/adapter/storage"
	"backend/src/infrastructure"
	"backend/src/usecase"

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

	secret := os.Getenv("API_GATEWAY_SECRET")

	pool, err := infrastructure.NewPool(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	log.Println("Running database migrations...")
	if err := runMigrations(databaseURL); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	repo := postgres.NewProductRepository(pool)
	inventoryRepo := postgres.NewInventoryRepository(pool)
	memberRepo := postgres.NewMemberRepository(pool)
	codeRepo := postgres.NewRegistrationCodeRepository(pool)
	categoryRepo := postgres.NewCategoryRepository(pool)
	sessionRepo := session.NewCache(time.Hour)
	productService := usecase.NewProductService(repo)
	inventoryService := usecase.NewInventoryService(inventoryRepo)
	memberService := usecase.NewMemberService(memberRepo, sessionRepo, codeRepo)
	sessionService := usecase.NewSessionService(sessionRepo)
	codeService := usecase.NewRegistrationCodeService(codeRepo)
	categoryService := usecase.NewCategoryService(categoryRepo)
	fileStorage := storage.LocalFileStorage{}
	announcementService := usecase.NewAnnouncementService(postgres.NewAnnouncementRepository(pool), fileStorage)
	chatService := usecase.NewChatService(postgres.NewChatRoomRepository(pool), fileStorage)
	eventService := usecase.NewEventService(postgres.NewEventRepository(pool))
	defer sessionRepo.Stop()

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/health", healthHandler).Methods("GET")

	apphttp.RegisterProductRoutes(r, productService, memberService, sessionService)
	apphttp.RegisterInventoryRoutes(r, inventoryService, memberService, sessionService)
	apphttp.RegisterMemberRoutes(r, memberService, sessionService, codeService)
	apphttp.RegisterRegistrationCodeRoutes(r, codeService, memberService, sessionService)
	apphttp.RegisterCategoryRoutes(r, categoryService, memberService, sessionService)
	apphttp.RegisterAnnouncementRoutes(r, announcementService, memberService, sessionService)
	apphttp.RegisterChatRoutes(r, chatService, memberService, sessionService)
	apphttp.RegisterEventRoutes(r, eventService, memberService, sessionService)

	handler := apphttp.GatewayMiddleware(secret)(r)
	if secret == "" {
		log.Println("API_GATEWAY_SECRET is not set — /api routes are open")
	}

	log.Println("Server starting on :8090")
	log.Fatal(http.ListenAndServe(":8090", handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "Welcome to Product Manager API"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "OK"})
}
