package multitenancy

import (
	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/platform/observability"
)

type Plugin struct {
	logger *observability.Logger
}

func (p *Plugin) Name() string {
	return "multitenancy"
}

func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, health *health.Health) error {
	p.logger = logger
	return nil
}

func (p *Plugin) Start() error {
	if p.logger != nil {
		p.logger.Info("Multitenancy Plugin (Stub) started")
	}
	return nil
}

func (p *Plugin) Stop() error {
	return nil
}
