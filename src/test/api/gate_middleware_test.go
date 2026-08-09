package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	api "backend/src/adapter/http"

	"github.com/stretchr/testify/assert"
)

func TestGatewayMiddleware(t *testing.T) {
	const secret = "super-secret"

	upstream := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	tests := []struct {
		name       string
		secret     string
		path       string
		authHeader string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "valid bearer passes",
			secret:     secret,
			path:       "/api/products",
			authHeader: "Bearer " + secret,
			wantStatus: http.StatusOK,
			wantBody:   "ok",
		},
		{
			name:       "missing header is rejected",
			secret:     secret,
			path:       "/api/products",
			wantStatus: http.StatusUnauthorized,
			wantBody:   "{\"error\":\"unauthorized\"}\n",
		},
		{
			name:       "wrong token is rejected",
			secret:     secret,
			path:       "/api/products",
			authHeader: "Bearer wrong-secret",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bare scheme is rejected",
			secret:     secret,
			path:       "/api/products",
			authHeader: "Bearer",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "non-bearer scheme is rejected",
			secret:     secret,
			path:       "/api/products",
			authHeader: "Basic x",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "health passes without header",
			secret:     secret,
			path:       "/api/health",
			wantStatus: http.StatusOK,
		},
		{
			name:       "media passes without header",
			secret:     secret,
			path:       "/media/images/announcements/foo.png",
			wantStatus: http.StatusOK,
		},
		{
			name:       "root passes without header",
			secret:     secret,
			path:       "/",
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty secret lets api pass without header",
			secret:     "",
			path:       "/api/products",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := api.GatewayMiddleware(tt.secret)(upstream)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantBody != "" {
				assert.Equal(t, tt.wantBody, w.Body.String())
			}
		})
	}
}
