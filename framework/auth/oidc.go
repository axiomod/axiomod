package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// OIDCConfig represents the configuration for the OIDC service
type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// OIDCDiscovery represents the OIDC discovery document
type OIDCDiscovery struct {
	Issuer      string `json:"issuer"`
	AuthURL     string `json:"authorization_endpoint"`
	TokenURL    string `json:"token_endpoint"`
	JWKSURL     string `json:"jwks_uri"`
	UserInfoURL string `json:"userinfo_endpoint"`
}

// OIDCService provides OIDC discovery and token verification
type OIDCService struct {
	config    OIDCConfig
	discovery *OIDCDiscovery
	mu        sync.RWMutex
}

// NewOIDCService creates a new OIDCService
func NewOIDCService(cfg OIDCConfig) *OIDCService {
	return &OIDCService{
		config: cfg,
	}
}

// Discover performs OIDC discovery
func (s *OIDCService) Discover(ctx context.Context) error {
	discoveryURL := fmt.Sprintf("%s/.well-known/openid-configuration", s.config.IssuerURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create discovery request: %w", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform discovery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("discovery failed with status: %s", resp.Status)
	}

	var discovery OIDCDiscovery
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return fmt.Errorf("failed to decode discovery document: %w", err)
	}

	s.mu.Lock()
	s.discovery = &discovery
	s.mu.Unlock()

	return nil
}

// VerifyToken verifies an OIDC ID token
func (s *OIDCService) VerifyToken(ctx context.Context, tokenString string) (*Claims, error) {
	s.mu.RLock()
	discovery := s.discovery
	s.mu.RUnlock()

	if discovery == nil {
		if err := s.Discover(ctx); err != nil {
			return nil, err
		}
		s.mu.RLock()
		discovery = s.discovery
		s.mu.RUnlock()
	}

	// For a real production implementation, we should fetch and cache JWKS from discovery.JWKSURL
	// and use those keys to verify the signature.
	// For this stabilization phase, we assume the token is signed with a method
	// that we can verify if we have the key, or we delegate to a more robust library.
	// Here we just parse the claims for now, but in a real app, this MUST verify the signature.

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// In a real implementation, you would:
		// 1. Get the 'kid' from the token header.
		// 2. Look up the public key from the cached JWKS.
		// 3. Return the public key.
		return nil, fmt.Errorf("signature verification not implementation in this version - use JWKS")
	})

	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token or claims")
	}

	return claims, nil
}
