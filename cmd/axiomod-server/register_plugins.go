package main

import (
	"github.com/axiomod/axiomod/plugins"
	"github.com/axiomod/axiomod/plugins/audit"
	"github.com/axiomod/axiomod/plugins/auth/ldap"
	"github.com/axiomod/axiomod/plugins/auth/saml"
	"github.com/axiomod/axiomod/plugins/cache/redis"
	"github.com/axiomod/axiomod/plugins/logging/elk"
	kafka_plugin "github.com/axiomod/axiomod/plugins/messaging/kafka"
	"github.com/axiomod/axiomod/plugins/middleware/multitenancy"
)

// RegisterNewPlugins registers the new decoupled plugins
func RegisterNewPlugins(r *plugins.PluginRegistry) error {
	r.Register(&ldap.Plugin{})
	r.Register(&saml.Plugin{})
	r.Register(&multitenancy.Plugin{})
	r.Register(&audit.Plugin{})
	r.Register(&elk.Plugin{})
	r.Register(&redis.Plugin{})
	r.Register(&kafka_plugin.Plugin{})
	return nil
}
