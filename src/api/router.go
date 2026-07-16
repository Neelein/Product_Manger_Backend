package api

import (
	"backend/src/domain"

	"github.com/gorilla/mux"
)

func RegisterMemberRoutes(r *mux.Router, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewMemberHandler(memberRepo, sessionRepo)

	r.HandleFunc("/api/members/register", h.RegisterMember).Methods("POST")
	r.HandleFunc("/api/members/login", h.LoginMember).Methods("POST")
	r.HandleFunc("/api/members/logout", h.LogoutMember).Methods("POST")
	r.HandleFunc("/api/members/me", h.GetCurrentMember).Methods("GET")
	r.HandleFunc("/api/members/update", h.UpdateMember).Methods("POST")
}

func RegisterProductRoutes(r *mux.Router, repo domain.ProductRepository) {
	h := NewProductHandler(repo)

	r.HandleFunc("/api/products", h.CreateProduct).Methods("POST")
	r.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	r.HandleFunc("/api/products/{id}", h.GetProduct).Methods("GET")
	r.HandleFunc("/api/products/{id}/update", h.UpdateProduct).Methods("POST")
	r.HandleFunc("/api/products/{id}/delete", h.DeleteProduct).Methods("POST")
}
