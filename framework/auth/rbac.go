package auth

import (
	"fmt"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/casbin/casbin/v2"
)

// RBACService provides role-based access control using Casbin
type RBACService struct {
	enforcer *casbin.Enforcer
}

// NewRBACService creates a new RBACService
func NewRBACService(cfg config.CasbinConfig) (*RBACService, error) {
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

// AddPolicy adds a policy to the enforcer
func (s *RBACService) AddPolicy(sub, obj, act string) (bool, error) {
	return s.enforcer.AddPolicy(sub, obj, act)
}

// RemovePolicy removes a policy from the enforcer
func (s *RBACService) RemovePolicy(sub, obj, act string) (bool, error) {
	return s.enforcer.RemovePolicy(sub, obj, act)
}

// AddRoleForUser adds a role for a user
func (s *RBACService) AddRoleForUser(user, role string) (bool, error) {
	return s.enforcer.AddGroupingPolicy(user, role)
}

// RemoveRoleForUser removes a role for a user
func (s *RBACService) RemoveRoleForUser(user, role string) (bool, error) {
	return s.enforcer.RemoveGroupingPolicy(user, role)
}

// GetRolesForUser returns all roles for a user
func (s *RBACService) GetRolesForUser(user string) ([]string, error) {
	return s.enforcer.GetRolesForUser(user)
}

// GetUsersForRole returns all users for a role
func (s *RBACService) GetUsersForRole(role string) ([]string, error) {
	return s.enforcer.GetUsersForRole(role)
}

// GetEnforcer returns the underlying casbin enforcer
func (s *RBACService) GetEnforcer() *casbin.Enforcer {
	return s.enforcer
}
