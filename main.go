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
	"backend/src/config"
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
	appConfig, err := config.Load(os.Getenv)
	if err != nil {
		log.Fatal(err)
	}
	databaseURL := appConfig.DatabaseURL
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		if err := runMigrations(databaseURL); err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		return
	}

	secret := appConfig.APIGatewaySecret

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
	fileStorage := storage.LocalFileStorage{}
	productService := usecase.NewProductService(repo, fileStorage)
	inventoryService := usecase.NewInventoryService(inventoryRepo)
	memberService := usecase.NewMemberService(memberRepo, sessionRepo, codeRepo)
	sessionService := usecase.NewSessionService(sessionRepo)
	codeService := usecase.NewRegistrationCodeService(codeRepo)
	categoryService := usecase.NewCategoryService(categoryRepo)
	departmentService := usecase.NewDepartmentService(memberRepo, memberRepo, postgres.NewDepartmentRepository(pool))
	announcementService := usecase.NewAnnouncementService(postgres.NewAnnouncementRepository(pool), fileStorage)
	chatService := usecase.NewChatService(postgres.NewChatRoomRepository(pool), fileStorage)
	eventService := usecase.NewEventService(postgres.NewEventRepository(pool))
	orderService := usecase.NewOrderService(postgres.NewOrderRepository(pool))
	paymentService := usecase.NewPaymentService(postgres.NewPaymentRepository(pool), appConfig.PaymentOTP)
	paymentWorker := usecase.NewPaymentWorker(paymentService, time.Now)
	workerContext, stopWorker := context.WithCancel(context.Background())
	defer stopWorker()
	go paymentWorker.Start(workerContext)
	defer sessionRepo.Stop()

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/health", healthHandler).Methods("GET")

	apphttp.RegisterProductRoutes(r, productService, memberService, sessionService)
	apphttp.RegisterInventoryRoutes(r, inventoryService, memberService, sessionService)
	apphttp.RegisterMemberRoutes(r, memberService, sessionService, codeService)
	apphttp.RegisterDepartmentRoutes(r, departmentService, memberService, sessionService, appConfig.DefaultDepartmentCode)
	apphttp.RegisterRegistrationCodeRoutes(r, codeService, memberService, sessionService)
	apphttp.RegisterCategoryRoutes(r, categoryService, memberService, sessionService)
	apphttp.RegisterAnnouncementRoutes(r, announcementService, memberService, sessionService)
	apphttp.RegisterChatRoutes(r, chatService, memberService, sessionService)
	apphttp.RegisterEventRoutes(r, eventService, memberService, sessionService)
	apphttp.RegisterOrderRoutes(r, orderService, memberService, sessionService)
	apphttp.RegisterPaymentRoutes(r, paymentService, memberService, sessionService)

	handler := apphttp.GatewayMiddleware(secret)(r)
	if secret == "" {
		log.Println("API_GATEWAY_SECRET is not set — /api routes are open")
	}

	log.Printf("Server starting on :%s", appConfig.Port)
	log.Fatal(http.ListenAndServe(":"+appConfig.Port, handler))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "Welcome to Product Manager API"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "OK"})
}
