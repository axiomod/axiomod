package example_plugin

// ExamplePlugin is an example plugin implementation
type ExamplePlugin struct {
	config map[string]interface{}
	active bool
}

// Name returns the name of the plugin
func (p *ExamplePlugin) Name() string {
	return "example"
}

// Initialize initializes the plugin with the given configuration
func (p *ExamplePlugin) Initialize(config map[string]interface{}) error {
	p.config = config
	return nil
}

// Start starts the plugin
func (p *ExamplePlugin) Start() error {
	p.active = true
	return nil
}

// Stop stops the plugin
func (p *ExamplePlugin) Stop() error {
	p.active = false
	return nil
}

// IsActive returns whether the plugin is active
func (p *ExamplePlugin) IsActive() bool {
	return p.active
}

// GetConfig returns the plugin configuration
func (p *ExamplePlugin) GetConfig() map[string]interface{} {
	return p.config
}
