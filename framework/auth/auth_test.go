package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTService(t *testing.T) {
	secret := "test-secret-key"
	duration := 1 * time.Hour
	service := NewJWTService(secret, duration)

	t.Run("Generate and Validate Token", func(t *testing.T) {
		userID := "user-123"
		username := "testuser"
		email := "test@example.com"
		roles := []string{"admin", "user"}

		token, err := service.GenerateToken(userID, username, email, roles)
		assert.NoError(t, err)
		assert.NotEmpty(t, token)

		claims, err := service.ValidateToken(token)
		assert.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, username, claims.Username)
		assert.Equal(t, email, claims.Email)
		assert.ElementsMatch(t, roles, claims.Roles)
	})

	t.Run("Expired Token", func(t *testing.T) {
		shortService := NewJWTService(secret, 1*time.Millisecond)
		token, err := shortService.GenerateToken("id", "user", "email", nil)
		assert.NoError(t, err)

		// Wait for expiration
		time.Sleep(2 * time.Millisecond)

		_, err = service.ValidateToken(token)
		assert.Error(t, err)
		assert.Equal(t, ErrExpiredToken, err)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		_, err := service.ValidateToken("not.a.token")
		assert.Error(t, err)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("Claims HasRole", func(t *testing.T) {
		claims := &Claims{Roles: []string{"admin", "editor"}}
		assert.True(t, claims.HasRole("admin"))
		assert.True(t, claims.HasRole("editor"))
		assert.False(t, claims.HasRole("viewer"))
	})
}

func TestOIDCService(t *testing.T) {
	cfg := OIDCConfig{
		IssuerURL: "https://example.com/auth/realms/master",
		ClientID:  "test-client",
	}
	service := NewOIDCService(cfg)

	t.Run("NewOIDCService", func(t *testing.T) {
		assert.NotNil(t, service)
		assert.Equal(t, cfg.IssuerURL, service.config.IssuerURL)
	})

	// Note: Fully testing Discovery and VerifyToken would require a mock OIDC server
	// which is complex for a unit test. For now, we verify the logic we can.

	t.Run("VerifyToken without Discovery", func(t *testing.T) {
		ctx := context.Background()
		// This should attempt discovery and fail because the URL is invalid in this environment
		_, err := service.VerifyToken(ctx, "invalid-token")
		assert.Error(t, err)
		// The error can be either a connection error or a 404/non-OK status depending on the environment
		assert.True(t, strings.Contains(err.Error(), "failed to perform discovery") || strings.Contains(err.Error(), "discovery failed with status"))
	})

	t.Run("Discovery with Mock Server", func(t *testing.T) {
		// Create a mock OIDC provider server
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Serve Discovery Document
			if strings.HasSuffix(r.URL.Path, "/.well-known/openid-configuration") {
				w.Header().Set("Content-Type", "application/json")
				// We need to use the server's actual URL for the endpoints
				_ = json.NewEncoder(w).Encode(OIDCDiscovery{
					Issuer:   "https://mock.com",           // Issuer can be anything as long as we validate it matching
					JWKSURL:  "http://" + r.Host + "/jwks", // Point back to this server
					AuthURL:  "http://" + r.Host + "/auth",
					TokenURL: "http://" + r.Host + "/token",
				})
				return
			}
			// Serve JWKS
			if strings.HasSuffix(r.URL.Path, "/jwks") {
				w.Header().Set("Content-Type", "application/json")
				// Return empty keyset for now, just to pass parsing
				w.Write([]byte(`{"keys": []}`))
				return
			}
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		mockCfg := OIDCConfig{
			IssuerURL: server.URL,
			ClientID:  "mock-client",
		}
		mockService := NewOIDCService(mockCfg)

		err := mockService.Discover(context.Background())
		// It will still fail on JWKS initialization because MicahParks/keyfunc tries to fetch the JWKS URL
		// but at least we cover the Discover method's first half.
		if err != nil {
			assert.True(t, strings.Contains(err.Error(), "failed to initialize JWKS"))
		}
		assert.NotNil(t, mockService.discovery)
		assert.Equal(t, "https://mock.com", mockService.discovery.Issuer)
	})
}
