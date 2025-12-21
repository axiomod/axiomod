package main

import (
	"github.com/axiomod/axiomod/plugins"
	"github.com/axiomod/axiomod/plugins/audit"
	"github.com/axiomod/axiomod/plugins/auth/ldap"
	"github.com/axiomod/axiomod/plugins/auth/saml"
	"github.com/axiomod/axiomod/plugins/logging/elk"
	"github.com/axiomod/axiomod/plugins/middleware/multitenancy"
)

// RegisterNewPlugins registers the new decoupled plugins
func RegisterNewPlugins(r *plugins.PluginRegistry) error {
	r.Register(&ldap.Plugin{})
	r.Register(&saml.Plugin{})
	r.Register(&multitenancy.Plugin{})
	r.Register(&audit.Plugin{})
	r.Register(&elk.Plugin{})
	return nil
}
