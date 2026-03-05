//go:build gherkin
// +build gherkin

package gherkin

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"begbot/internal/auth"

	"github.com/cucumber/godog"
	"github.com/golang-jwt/jwt/v5"
)

type authTestState struct {
	middleware  *auth.AuthMiddleware
	response    *httptest.ResponseRecorder
	validToken  string
	jwksServer  *httptest.Server
	userIDInCtx string
}

// InitializeAuthenticationScenario registers BDD step definitions for authentication tests.
func InitializeAuthenticationScenario(ctx *godog.ScenarioContext) {
	state := &authTestState{}

	ctx.BeforeScenario(func(sc *godog.Scenario) {
		state = &authTestState{}
		state.middleware = auth.NewAuthMiddleware("https://test.supabase.co", "")
	})

	ctx.AfterScenario(func(sc *godog.Scenario, err error) {
		if state.jwksServer != nil {
			state.jwksServer.Close()
			state.jwksServer = nil
		}
	})

	// Background
	ctx.Given(`^an authentication middleware is configured$`, func() error {
		state.middleware = auth.NewAuthMiddleware("https://test.supabase.co", "")
		return nil
	})

	// When: no Authorization header at all
	ctx.When(`^a GET request is made to "([^"]*)" without an Authorization header$`, func(path string) error {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		state.response = httptest.NewRecorder()
		makeAuthHandler(state).ServeHTTP(state.response, req)
		return nil
	})

	// When: request with a specific Authorization header value
	ctx.When(`^a GET request is made to "([^"]*)" with Authorization header "([^"]*)"$`, func(path, headerValue string) error {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", headerValue)
		state.response = httptest.NewRecorder()
		makeAuthHandler(state).ServeHTTP(state.response, req)
		return nil
	})

	// When: request using the valid token set up in a Given step
	ctx.When(`^a GET request is made to "([^"]*)" with the valid Bearer token$`, func(path string) error {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer "+state.validToken)
		state.response = httptest.NewRecorder()
		makeAuthHandler(state).ServeHTTP(state.response, req)
		return nil
	})

	// Given: set up a valid JWT token backed by a local JWKS server
	ctx.Given(`^a valid JWT token signed with a test RSA key$`, func() error {
		// Generate a throwaway RSA key pair
		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		if err != nil {
			return fmt.Errorf("failed to generate RSA key pair: %w", err)
		}

		const testKid = "test-key-id"

		// Encode the public key as a JWKS entry
		pubKey := &privateKey.PublicKey
		nBytes := pubKey.N.Bytes()
		eVal := pubKey.E
		// Convert exponent to minimal big-endian bytes (typically AQAB for 65537)
		eBytes := []byte{byte(eVal >> 16), byte(eVal >> 8), byte(eVal)}
		for len(eBytes) > 1 && eBytes[0] == 0 {
			eBytes = eBytes[1:]
		}

		jwksBody := map[string]interface{}{
			"keys": []map[string]interface{}{
				{
					"kid": testKid,
					"kty": "RSA",
					"alg": "RS256",
					"use": "sig",
					"n":   base64.RawURLEncoding.EncodeToString(nBytes),
					"e":   base64.RawURLEncoding.EncodeToString(eBytes),
				},
			},
		}

		// Start a local JWKS server so the middleware can fetch the public key
		state.jwksServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(jwksBody) //nolint:errcheck
		}))

		// Point the middleware at the local JWKS server
		state.middleware = auth.NewAuthMiddleware(state.jwksServer.URL, "")

		// Create and sign a JWT using the matching private key
		now := time.Now()
		claims := jwt.MapClaims{
			"sub":   "user-abc-123",
			"email": "test@example.com",
			"role":  "authenticated",
			"iat":   now.Unix(),
			"exp":   now.Add(time.Hour).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
		token.Header["kid"] = testKid

		state.validToken, err = token.SignedString(privateKey)
		if err != nil {
			return fmt.Errorf("failed to sign JWT: %w", err)
		}
		return nil
	})

	// Then: assert HTTP status code
	ctx.Then(`^the response status should be (\d+)$`, func(expected int) error {
		if state.response.Code != expected {
			return fmt.Errorf("expected HTTP status %d, got %d", expected, state.response.Code)
		}
		return nil
	})

	// Then (And): assert the JSON body contains a specific error code
	ctx.Then(`^the response should contain error code "([^"]*)"$`, func(expectedCode string) error {
		var body map[string]interface{}
		if err := json.Unmarshal(state.response.Body.Bytes(), &body); err != nil {
			return fmt.Errorf("failed to parse response body as JSON: %w", err)
		}
		errField, ok := body["error"]
		if !ok {
			return fmt.Errorf("expected 'error' field in response body, got: %s", state.response.Body.String())
		}
		errMap, ok := errField.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected 'error' to be a JSON object")
		}
		code, _ := errMap["code"].(string)
		if code != expectedCode {
			return fmt.Errorf("expected error code %q, got %q", expectedCode, code)
		}
		return nil
	})

	// Then (And): assert user ID was extracted from the JWT and placed in context
	ctx.Then(`^the user ID should be set in the request context$`, func() error {
		if state.userIDInCtx == "" {
			return fmt.Errorf("expected user ID in request context, but it was empty")
		}
		return nil
	})
}

// makeAuthHandler wraps the next handler to capture the user ID injected into the context.
func makeAuthHandler(state *authTestState) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _ := auth.GetUserID(r.Context())
		state.userIDInCtx = userID
		w.WriteHeader(http.StatusOK)
	})
	return state.middleware.Middleware(next)
}

// TestAuthenticationFeatures runs the Godog authentication BDD tests.
func TestAuthenticationFeatures(t *testing.T) {
	featurePath := getFeaturesPath("authentication.feature")
	suite := godog.TestSuite{
		ScenarioInitializer: InitializeAuthenticationScenario,
		Options: &godog.Options{
			Format:   "pretty",
			Paths:    []string{featurePath},
			TestingT: t,
		},
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run authentication gherkin tests")
	}
}
