package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/axiomod/axiomod/platform/observability"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
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
	config        OIDCConfig
	discovery     *OIDCDiscovery
	jwks          keyfunc.Keyfunc
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	logger        *observability.Logger
	lastDiscovery time.Time
}

// NewOIDCService creates a new OIDCService
func NewOIDCService(cfg OIDCConfig, logger *observability.Logger) *OIDCService {
	if cfg.JWKSCacheTTL == 0 {
		cfg.JWKSCacheTTL = 1 * time.Hour
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &OIDCService{
		config: cfg,
		ctx:    ctx,
		cancel: cancel,
		logger: logger,
	}
}

// Start initiates the background refresh of discovery and JWKS
func (s *OIDCService) Start() {
	// Initial discovery
	if err := s.Discover(s.ctx); err != nil {
		s.logger.Error("Initial OIDC discovery failed", zap.Error(err))
	}

	// Start background refresh ticker
	go func() {
		ticker := time.NewTicker(s.config.JWKSCacheTTL)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := s.Discover(s.ctx); err != nil {
					s.logger.Error("Background OIDC discovery failed", zap.Error(err))
				}
			case <-s.ctx.Done():
				return
			}
		}
	}()
}

// Stop stops the background refresh
func (s *OIDCService) Stop() {
	s.cancel()
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
	s.lastDiscovery = time.Now()
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
			return nil, fmt.Errorf("OIDC discovery failed and no cached JWKS: %w", err)
		}
		s.mu.RLock()
		discovery = s.discovery
		jwks = s.jwks
		s.mu.RUnlock()
	}

	// Check for stale discovery (e.g. older than 2x TTL)
	s.mu.RLock()
	lastDisco := s.lastDiscovery
	s.mu.RUnlock()

	if time.Since(lastDisco) > s.config.JWKSCacheTTL*2 && lastDisco.IsZero() == false {
		s.logger.Warn("OIDC discovery is stale", zap.Time("last_success", lastDisco))
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
