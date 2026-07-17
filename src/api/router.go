package api

import (
	"net/http"

	"backend/src/domain"

	"github.com/gorilla/mux"
)

func RegisterProductRoutes(r *mux.Router, repo domain.ProductRepository, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewProductHandler(repo)
	auth := AuthMiddleware(sessionRepo, memberRepo)

	r.HandleFunc("/api/products", h.ListProducts).Methods("GET")
	r.Handle("/api/products", auth(http.HandlerFunc(h.CreateProduct))).Methods("POST")
	r.HandleFunc("/api/products/{id}", h.GetProduct).Methods("GET")
	r.HandleFunc("/api/products/{id}/update", h.UpdateProduct).Methods("POST")
	r.HandleFunc("/api/products/{id}/delete", h.DeleteProduct).Methods("POST")
	r.HandleFunc("/api/products/{id}/detail", h.GetDetail).Methods("GET")
	r.Handle("/api/products/{id}/detail/update", auth(http.HandlerFunc(h.UpdateDetail))).Methods("POST")
	r.HandleFunc("/api/products/{id}/detail/prices", h.ListPrices).Methods("GET")
	r.HandleFunc("/api/products/{id}/detail/prices/{pid}", h.GetPrice).Methods("GET")
	r.Handle("/api/products/{id}/detail/prices/{pid}/update", auth(http.HandlerFunc(h.UpdatePrice))).Methods("POST")
	r.Handle("/api/products/{id}/details", auth(http.HandlerFunc(h.CreateDetail))).Methods("POST")
	r.Handle("/api/products/{id}/details/{did}/prices", auth(http.HandlerFunc(h.CreatePrice))).Methods("POST")
}

func RegisterMemberRoutes(r *mux.Router, memberRepo domain.MemberRepository, sessionRepo domain.SessionRepository) {
	h := NewMemberHandler(memberRepo, sessionRepo)
	auth := AuthMiddleware(sessionRepo, memberRepo)

	r.HandleFunc("/api/members/register", h.RegisterMember).Methods("POST")
	r.HandleFunc("/api/members/login", h.LoginMember).Methods("POST")
	r.HandleFunc("/api/members/logout", h.LogoutMember).Methods("POST")
	r.Handle("/api/members/me", auth(http.HandlerFunc(h.GetCurrentMember))).Methods("GET")
	r.Handle("/api/members/update", auth(http.HandlerFunc(h.UpdateMember))).Methods("POST")
}
