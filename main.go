package main

import (
	"encoding/json"
	"log"
	"net/http"

	"backend/src/api"
	"backend/src/database"

	"github.com/gorilla/mux"
)

type Response struct {
	Message string `json:"message"`
}

func main() {
	repo := database.NewInMemoryRepository()

	r := mux.NewRouter()

	r.HandleFunc("/", homeHandler).Methods("GET")
	r.HandleFunc("/api/health", healthHandler).Methods("GET")

	api.RegisterProductRoutes(r, repo)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "Welcome to Product Manager API"})
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(Response{Message: "OK"})
}
