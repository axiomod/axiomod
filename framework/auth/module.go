package auth

import (
	"context"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/platform/observability"
	"go.uber.org/fx"
)

// Module provides the fx options for the auth module
var Module = fx.Options(
	fx.Provide(ProvideJWTService),
	fx.Provide(ProvideOIDCService),
	fx.Provide(ProvideRBACService),
	fx.Invoke(RegisterOIDCLifecycle),
)

// ProvideJWTService provides a JWTService
func ProvideJWTService(cfg *config.Config) *JWTService {
	return NewJWTService(
		cfg.Auth.JWT.SecretKey,
		time.Duration(cfg.Auth.JWT.TokenDuration)*time.Minute,
	)
}

// ProvideOIDCService provides an OIDCService
func ProvideOIDCService(cfg *config.Config, logger *observability.Logger) *OIDCService {
	return NewOIDCService(OIDCConfig{
		IssuerURL:    cfg.Auth.OIDC.IssuerURL,
		ClientID:     cfg.Auth.OIDC.ClientID,
		ClientSecret: cfg.Auth.OIDC.ClientSecret,
		JWKSCacheTTL: time.Duration(cfg.Auth.OIDC.JWKSCacheTTL) * time.Minute,
	}, logger)
}

// ProvideRBACService provides an RBACService
func ProvideRBACService(cfg *config.Config) (*RBACService, error) {
	return NewRBACService(cfg.Casbin)
}

// RegisterOIDCLifecycle registers the OIDCService with the fx lifecycle
func RegisterOIDCLifecycle(lc fx.Lifecycle, s *OIDCService) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			s.Start()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.Stop()
			return nil
		},
	})
}
