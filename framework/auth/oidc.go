package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
)

// OIDCConfig represents the configuration for the OIDC service
type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	JWKSCacheTTL time.Duration
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
	jwks      keyfunc.Keyfunc
	mu        sync.RWMutex
}

// NewOIDCService creates a new OIDCService
func NewOIDCService(cfg OIDCConfig) *OIDCService {
	if cfg.JWKSCacheTTL == 0 {
		cfg.JWKSCacheTTL = 1 * time.Hour
	}
	return &OIDCService{
		config: cfg,
	}
}

// Discover performs OIDC discovery and initializes JWKS
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

	// Initialize JWKS
	kf, err := keyfunc.NewDefault([]string{discovery.JWKSURL})
	if err != nil {
		return fmt.Errorf("failed to initialize JWKS: %w", err)
	}

	s.mu.Lock()
	s.discovery = &discovery
	s.jwks = kf
	s.mu.Unlock()

	return nil
}

// VerifyToken verifies an OIDC ID token
func (s *OIDCService) VerifyToken(ctx context.Context, tokenString string) (*Claims, error) {
	s.mu.RLock()
	discovery := s.discovery
	jwks := s.jwks
	s.mu.RUnlock()

	if jwks == nil {
		if err := s.Discover(ctx); err != nil {
			return nil, err
		}
		s.mu.RLock()
		discovery = s.discovery
		jwks = s.jwks
		s.mu.RUnlock()
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, jwks.Keyfunc)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid token or claims")
	}

	// Verify issuer and audience
	if claims.Issuer != discovery.Issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", discovery.Issuer, claims.Issuer)
	}

	// Aud usually contains ClientID for ID tokens
	aud, err := claims.GetAudience()
	if err != nil {
		return nil, fmt.Errorf("failed to get audience: %w", err)
	}

	foundAud := false
	for _, a := range aud {
		if a == s.config.ClientID {
			foundAud = true
			break
		}
	}
	if !foundAud {
		return nil, fmt.Errorf("invalid audience: %v", aud)
	}

	return claims, nil
}
