package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test that unauthorized requests are rejected
func TestAuthMiddleware_Unauthorized(t *testing.T) {
	authMiddleware := NewAuthMiddleware("https://fxhknzpqqhrkpqothjvrx.supabase.co")

	handler := authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// Test without Authorization header
	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// Test that requests with invalid Bearer token are rejected
func TestAuthMiddleware_InvalidToken(t *testing.T) {
	authMiddleware := NewAuthMiddleware("https://fxhknzpqqhrkpqothjvrx.supabase.co")

	handler := authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// Test with invalid token
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

// Test that requests without Bearer prefix are rejected
func TestAuthMiddleware_MissingBearerPrefix(t *testing.T) {
	authMiddleware := NewAuthMiddleware("https://fxhknzpqqhrkpqothjvrx.supabase.co")

	handler := authMiddleware.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	}))

	// Test without Bearer prefix
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "some-token")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}
