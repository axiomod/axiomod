package auth

import (
	"fmt"

	"github.com/casbin/casbin/v2"
)

// CasbinConfig represents the configuration for Casbin
type CasbinConfig struct {
	ModelPath  string
	PolicyPath string
	Table      string // for database policy (if implemented later)
}

// RBACService provides role-based access control using Casbin
type RBACService struct {
	enforcer *casbin.Enforcer
}

// NewRBACService creates a new RBACService
func NewRBACService(cfg CasbinConfig) (*RBACService, error) {
	enforcer, err := casbin.NewEnforcer(cfg.ModelPath, cfg.PolicyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create casbin enforcer: %w", err)
	}

	return &RBACService{
		enforcer: enforcer,
	}, nil
}

// Enforce checks if a subject can perform an action on a resource
func (s *RBACService) Enforce(sub, obj, act string) (bool, error) {
	return s.enforcer.Enforce(sub, obj, act)
}

// ReloadPolicy reloads the policy from storage
func (s *RBACService) ReloadPolicy() error {
	return s.enforcer.LoadPolicy()
}
