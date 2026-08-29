package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequireEmployeeRejectsCustomer(t *testing.T) {
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			member := &Member{MemberType: "customer"}
			next.ServeHTTP(w, r.WithContext(ContextWithMember(context.Background(), member)))
		})
	}
	handler := RequireEmployee(auth)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { t.Fatal("customer reached employee route") }))
	req := httptest.NewRequest(http.MethodPost, "/api/products", nil)
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
}

func TestRequireRoleUsesPermissionOnly(t *testing.T) {
	auth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			member := &Member{MemberType: "employee"}
			next.ServeHTTP(w, r.WithContext(ContextWithMember(context.Background(), member)))
		})
	}
	handler := RequireRole("admin", auth)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("employee without permission reached admin route")
	}))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequest(http.MethodPost, "/api/registration-codes", nil))
	if res.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", res.Code)
	}
}
