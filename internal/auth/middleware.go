package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// JWK represents a JSON Web Key
type JWK struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// SupabaseClaims represents the claims in a Supabase JWT token
type SupabaseClaims struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Role  string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware handles Supabase JWT authentication
type AuthMiddleware struct {
	supabaseURL string
	jwksURL     string
	publicKeys  map[string]*rsa.PublicKey
	mu          sync.RWMutex
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(supabaseURL string) *AuthMiddleware {
	if supabaseURL == "" {
		supabaseURL = "https://fxhknzpqqhrkpqothjvrx.supabase.co"
	}
	
	return &AuthMiddleware{
		supabaseURL: supabaseURL,
		jwksURL:     fmt.Sprintf("%s/auth/v1/jwks", supabaseURL),
		publicKeys:  make(map[string]*rsa.PublicKey),
	}
}

// Middleware wraps HTTP handlers with authentication
func (am *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "Missing authorization header")
			return
		}

		// Check for Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			writeUnauthorized(w, "Invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := am.ValidateToken(tokenString)
		if err != nil {
			writeUnauthorized(w, fmt.Sprintf("Invalid token: %v", err))
			return
		}

		// Add user ID to context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.Sub)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ValidateToken validates a JWT token and returns the claims
func (am *AuthMiddleware) ValidateToken(tokenString string) (*SupabaseClaims, error) {
	// Parse token without verification first to get the kid
	token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Get kid from token header
		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		// Get public key for this kid
		publicKey, err := am.getPublicKey(kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*SupabaseClaims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	return claims, nil
}

// getPublicKey retrieves the public key for a given kid
func (am *AuthMiddleware) getPublicKey(kid string) (*rsa.PublicKey, error) {
	// Check if we already have this key
	am.mu.RLock()
	if key, exists := am.publicKeys[kid]; exists {
		am.mu.RUnlock()
		return key, nil
	}
	am.mu.RUnlock()

	// Fetch JWKS
	am.mu.Lock()
	defer am.mu.Unlock()

	// Double-check after acquiring write lock
	if key, exists := am.publicKeys[kid]; exists {
		return key, nil
	}

	// Fetch JWKS from Supabase
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(am.jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch JWKS: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read JWKS response: %w", err)
	}

	var jwks JWKS
	if err := json.Unmarshal(body, &jwks); err != nil {
		return nil, fmt.Errorf("failed to parse JWKS: %w", err)
	}

	// Find the key with matching kid
	var targetJWK *JWK
	for i, key := range jwks.Keys {
		if key.Kid == kid {
			targetJWK = &jwks.Keys[i]
			break
		}
	}

	if targetJWK == nil {
		return nil, fmt.Errorf("key with kid %s not found in JWKS", kid)
	}

	// Convert JWK to RSA public key
	publicKey, err := jwkToRSAPublicKey(targetJWK)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JWK to RSA public key: %w", err)
	}

	// Cache the key
	am.publicKeys[kid] = publicKey

	return publicKey, nil
}

// jwkToRSAPublicKey converts a JWK to an RSA public key
func jwkToRSAPublicKey(jwk *JWK) (*rsa.PublicKey, error) {
	// Decode N (modulus)
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode N: %w", err)
	}

	// Decode E (exponent)
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode E: %w", err)
	}

	// Convert to big.Int
	n := new(big.Int).SetBytes(nBytes)
	
	// Convert exponent bytes to int
	var e int
	for _, b := range eBytes {
		e = e<<8 | int(b)
	}

	return &rsa.PublicKey{
		N: n,
		E: e,
	}, nil
}

// writeUnauthorized writes a 401 Unauthorized response
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"message": message,
			"code":    "UNAUTHORIZED",
		},
	})
}

// GetUserID extracts the user ID from the request context
func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}
