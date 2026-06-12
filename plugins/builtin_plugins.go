package plugins

import (
	"context"
	"fmt"
	"time"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/database"
	"github.com/axiomod/axiomod/framework/health"

	"github.com/axiomod/axiomod/platform/observability"
	"go.uber.org/zap"
)

// MySQLPlugin implements the MySQL database plugin
type MySQLPlugin struct {
	config  map[string]interface{}
	db      *database.DB
	logger  *observability.Logger
	metrics *observability.Metrics
	health  *health.Health
	cfg     *config.Config
}

// Name returns the name of the plugin
func (p *MySQLPlugin) Name() string {
	return "mysql"
}

// Initialize initializes the plugin with the given configuration, logger, and metrics
func (p *MySQLPlugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.config = settings
	p.logger = logger
	p.metrics = metrics
	p.health = health
	p.cfg = cfg
	return nil
}

// Start starts the plugin
func (p *MySQLPlugin) Start() error {
	// Connect to the database using the simplified Connect method
	db, err := database.Connect(p.cfg, p.logger, p.metrics, p.health)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

// Stop stops the plugin
func (p *MySQLPlugin) Stop() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// PostgreSQLPlugin implements the PostgreSQL database plugin
type PostgreSQLPlugin struct {
	config  map[string]interface{}
	db      *database.DB
	logger  *observability.Logger
	metrics *observability.Metrics
	health  *health.Health
	cfg     *config.Config
}

// Name returns the name of the plugin
func (p *PostgreSQLPlugin) Name() string {
	return "postgresql"
}

// Initialize initializes the plugin with the given configuration, logger, and metrics
func (p *PostgreSQLPlugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.config = settings
	p.logger = logger
	p.metrics = metrics
	p.health = health
	p.cfg = cfg
	return nil
}

// Start starts the plugin
func (p *PostgreSQLPlugin) Start() error {
	// Connect to the database using the simplified Connect method
	db, err := database.Connect(p.cfg, p.logger, p.metrics, p.health)
	if err != nil {
		return err
	}
	p.db = db
	return nil
}

// Stop stops the plugin
func (p *PostgreSQLPlugin) Stop() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// JWTPlugin implements the JWT authentication plugin
type JWTPlugin struct {
	config  map[string]interface{}
	service *auth.JWTService
	logger  *observability.Logger
	metrics *observability.Metrics
	cfg     *config.Config
}

// Name returns the name of the plugin
func (p *JWTPlugin) Name() string {
	return "jwt"
}

// Initialize initializes the plugin with the given configuration, logger, and metrics
func (p *JWTPlugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.config = settings
	p.logger = logger
	p.metrics = metrics
	p.cfg = cfg
	return nil
}

// Start starts the plugin
func (p *JWTPlugin) Start() error {
	secret, ok := p.config["secret"].(string)
	if !ok || secret == "" {
		// Fall back to the central auth config if the plugin settings omit it.
		secret = p.cfg.Auth.JWT.SecretKey
	}
	if err := auth.ValidateSecretKey(secret); err != nil {
		return fmt.Errorf("jwt plugin: %w", err)
	}

	duration := 24 * time.Hour
	if durationStr, ok := p.config["duration"].(string); ok && durationStr != "" {
		parsed, err := time.ParseDuration(durationStr)
		if err != nil {
			return fmt.Errorf("jwt plugin: invalid duration %q: %w", durationStr, err)
		}
		duration = parsed
	}

	p.service = auth.NewJWTService(secret, duration)
	p.logger.Info("JWT service initialized")
	return nil
}

// Stop stops the plugin
func (p *JWTPlugin) Stop() error {
	return nil
}

// KeycloakPlugin implements the Keycloak authentication plugin
type KeycloakPlugin struct {
	config         map[string]interface{}
	service        *auth.OIDCService
	logger         *observability.Logger
	metrics        *observability.Metrics
	cfg            *config.Config
	discoverCancel context.CancelFunc
	discoverDone   chan struct{}
}

// Name returns the name of the plugin
func (p *KeycloakPlugin) Name() string {
	return "keycloak"
}

// Initialize initializes the plugin with the given configuration, logger, and metrics
func (p *KeycloakPlugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.config = settings
	p.logger = logger
	p.metrics = metrics
	p.cfg = cfg
	return nil
}

// Start starts the plugin
func (p *KeycloakPlugin) Start() error {
	issuer, _ := p.config["issuer"].(string)
	clientID, _ := p.config["client_id"].(string)
	clientSecret, _ := p.config["client_secret"].(string)

	if issuer == "" {
		return fmt.Errorf("keycloak issuer URL is required")
	}

	p.service = auth.NewOIDCService(auth.OIDCConfig{
		IssuerURL:    issuer,
		ClientID:     clientID,
		ClientSecret: clientSecret,
	}, p.logger)

	// Perform discovery in the background to avoid blocking startup if
	// Keycloak is down. The goroutine is cancellable from Stop().
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	p.discoverCancel = cancel
	p.discoverDone = make(chan struct{})
	go func() {
		defer cancel()
		defer close(p.discoverDone)
		if err := p.service.Discover(ctx); err != nil {
			p.logger.Error("Failed to discover Keycloak OIDC configuration", zap.Error(err))
		} else {
			p.logger.Info("Keycloak OIDC discovery successful")
		}
	}()

	return nil
}

// Stop stops the plugin, cancelling any in-flight discovery and the
// OIDC service's background refresh loop.
func (p *KeycloakPlugin) Stop() error {
	if p.discoverCancel != nil {
		p.discoverCancel()
		<-p.discoverDone
	}
	if p.service != nil {
		p.service.Stop()
	}
	return nil
}

// CasdoorPlugin implements the Casdoor authentication plugin
type CasdoorPlugin struct {
	config  map[string]interface{}
	logger  *observability.Logger
	metrics *observability.Metrics
	cfg     *config.Config
}

// Name returns the name of the plugin
func (p *CasdoorPlugin) Name() string {
	return "casdoor"
}

// Initialize initializes the plugin with the given configuration, logger, and metrics
func (p *CasdoorPlugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.config = settings
	p.logger = logger
	p.metrics = metrics
	p.cfg = cfg
	return nil
}

// Start starts the plugin
func (p *CasdoorPlugin) Start() error {
	return nil
}

// Stop stops the plugin
func (p *CasdoorPlugin) Stop() error {
	return nil
}

// CasbinPlugin implements the Casbin authorization plugin
type CasbinPlugin struct {
	config  map[string]interface{}
	logger  *observability.Logger
	metrics *observability.Metrics
	cfg     *config.Config
}

// Name returns the name of the plugin
func (p *CasbinPlugin) Name() string {
	return "casbin"
}

// Initialize initializes the plugin with the given configuration, logger, and metrics
func (p *CasbinPlugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.config = settings
	p.logger = logger
	p.metrics = metrics
	p.cfg = cfg
	return nil
}

// Start starts the plugin
func (p *CasbinPlugin) Start() error {
	return nil
}

// Stop stops the plugin
func (p *CasbinPlugin) Stop() error {
	return nil
}
