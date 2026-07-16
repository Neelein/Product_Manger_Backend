package api

import (
	"backend/src/domain"

	"github.com/gorilla/mux"
)

func RegisterProductRoutes(r *mux.Router, repo domain.ProductRepository) {
	h := NewProductHandler(repo)

	r.HandleFunc("/api/products", h.CreateProduct).Methods("POST")
	r.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	r.HandleFunc("/api/products/{id}", h.GetProduct).Methods("GET")
	r.HandleFunc("/api/products/{id}/update", h.UpdateProduct).Methods("POST")
	r.HandleFunc("/api/products/{id}/delete", h.DeleteProduct).Methods("POST")
}
