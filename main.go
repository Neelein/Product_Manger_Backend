package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"backend/src/api"
	"backend/src/database"

	"github.com/gorilla/mux"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	pool, err := database.NewPool(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()
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

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "Welcome to Product Manager API"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "OK"})
}
