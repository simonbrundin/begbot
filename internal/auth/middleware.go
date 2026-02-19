package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
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
	Kid   string `json:"kid"`
	Kty   string `json:"kty"`
	Alg   string `json:"alg"`
	Use   string `json:"use"`
	N     string `json:"n"`
	E     string `json:"e"`
	X     string `json:"x"`
	Y     string `json:"y"`
	Curve string `json:"crv"`
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
	supabaseURL     string
	supabaseAnonKey string
	jwksURL         string
	rsaPublicKeys   map[string]*rsa.PublicKey
	ecdsaPublicKeys map[string]*ecdsa.PublicKey
	mu              sync.RWMutex
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(supabaseURL, supabaseAnonKey string) *AuthMiddleware {
	if supabaseURL == "" {
		supabaseURL = "https://fxhknzpqhrkpqothjvrx.supabase.co"
	}

	return &AuthMiddleware{
		supabaseURL:     supabaseURL,
		supabaseAnonKey: supabaseAnonKey,
		jwksURL:         fmt.Sprintf("%s/auth/v1/.well-known/jwks.json", supabaseURL),
		rsaPublicKeys:   make(map[string]*rsa.PublicKey),
		ecdsaPublicKeys: make(map[string]*ecdsa.PublicKey),
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
	var signingMethod jwt.SigningMethod

	token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
		signingMethod = token.Method

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing kid in token header")
		}

		switch token.Method {
		case jwt.SigningMethodRS256, jwt.SigningMethodRS384, jwt.SigningMethodRS512:
			return am.getRSAPublicKey(kid)
		case jwt.SigningMethodES256, jwt.SigningMethodES384, jwt.SigningMethodES512:
			return am.getECDSAPublicKey(kid)
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
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

	_ = signingMethod
	return claims, nil
}

// getRSAPublicKey retrieves the RSA public key for a given kid
func (am *AuthMiddleware) getRSAPublicKey(kid string) (*rsa.PublicKey, error) {
	am.mu.RLock()
	if key, exists := am.rsaPublicKeys[kid]; exists {
		am.mu.RUnlock()
		return key, nil
	}
	am.mu.RUnlock()

	am.mu.Lock()
	defer am.mu.Unlock()

	if key, exists := am.rsaPublicKeys[kid]; exists {
		return key, nil
	}

	keys, err := am.fetchJWKS()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.Kid == kid && key.Kty == "RSA" {
			publicKey, err := jwkToRSAPublicKey(&key)
			if err != nil {
				return nil, err
			}
			am.rsaPublicKeys[kid] = publicKey
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("RSA key with kid %s not found in JWKS", kid)
}

// getECDSAPublicKey retrieves the ECDSA public key for a given kid
func (am *AuthMiddleware) getECDSAPublicKey(kid string) (*ecdsa.PublicKey, error) {
	am.mu.RLock()
	if key, exists := am.ecdsaPublicKeys[kid]; exists {
		am.mu.RUnlock()
		return key, nil
	}
	am.mu.RUnlock()

	am.mu.Lock()
	defer am.mu.Unlock()

	if key, exists := am.ecdsaPublicKeys[kid]; exists {
		return key, nil
	}

	keys, err := am.fetchJWKS()
	if err != nil {
		return nil, err
	}

	for _, key := range keys {
		if key.Kid == kid && key.Kty == "EC" {
			publicKey, err := jwkToECDSAPublicKey(&key)
			if err != nil {
				return nil, err
			}
			am.ecdsaPublicKeys[kid] = publicKey
			return publicKey, nil
		}
	}

	return nil, fmt.Errorf("EC key with kid %s not found in JWKS", kid)
}

// fetchJWKS fetches and parses the JWKS from Supabase
func (am *AuthMiddleware) fetchJWKS() ([]JWK, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", am.jwksURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS request: %w", err)
	}
	req.Header.Set("apikey", am.supabaseAnonKey)
	resp, err := client.Do(req)
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

	return jwks.Keys, nil
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

// jwkToECDSAPublicKey converts a JWK to an ECDSA public key
func jwkToECDSAPublicKey(jwk *JWK) (*ecdsa.PublicKey, error) {
	xBytes, err := base64.RawURLEncoding.DecodeString(jwk.X)
	if err != nil {
		return nil, fmt.Errorf("failed to decode X: %w", err)
	}

	yBytes, err := base64.RawURLEncoding.DecodeString(jwk.Y)
	if err != nil {
		return nil, fmt.Errorf("failed to decode Y: %w", err)
	}

	var curve elliptic.Curve
	switch jwk.Curve {
	case "P-256":
		curve = elliptic.P256()
	case "P-384":
		curve = elliptic.P384()
	case "P-521":
		curve = elliptic.P521()
	default:
		return nil, fmt.Errorf("unsupported curve: %s", jwk.Curve)
	}

	return &ecdsa.PublicKey{
		Curve: curve,
		X:     new(big.Int).SetBytes(xBytes),
		Y:     new(big.Int).SetBytes(yBytes),
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
