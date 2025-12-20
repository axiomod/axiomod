package config

// PluginsConfig represents the plugins configuration
type PluginsConfig struct {
	Enabled map[string]bool
	Settings map[string]map[string]interface{}
	Paths []string
}
