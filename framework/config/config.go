package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Provider defines the interface for configuration providers
type Provider interface {
	// Get retrieves a value from the configuration
	Get(key string) interface{}

	// GetString retrieves a string value from the configuration
	GetString(key string) string

	// GetInt retrieves an integer value from the configuration
	GetInt(key string) int

	// GetBool retrieves a boolean value from the configuration
	GetBool(key string) bool

	// GetFloat64 retrieves a float64 value from the configuration
	GetFloat64(key string) float64

	// GetDuration retrieves a duration value from the configuration
	GetDuration(key string) time.Duration

	// GetStringSlice retrieves a string slice from the configuration
	GetStringSlice(key string) []string

	// GetStringMap retrieves a string map from the configuration
	GetStringMap(key string) map[string]interface{}

	// GetStringMapString retrieves a string map of strings from the configuration
	GetStringMapString(key string) map[string]string

	// IsSet checks if a key is set in the configuration
	IsSet(key string) bool

	// AllSettings returns all settings from the configuration
	AllSettings() map[string]interface{}

	// WatchConfig watches for changes in the configuration
	WatchConfig(onChange func())
}

// ViperProvider implements the Provider interface using Viper
type ViperProvider struct {
	viper *viper.Viper
}

// WatchConfig watches for changes in the configuration
func (p *ViperProvider) WatchConfig(onChange func()) {
	p.viper.WatchConfig()
	p.viper.OnConfigChange(func(e fsnotify.Event) {
		onChange()
	})
}

// NewViperProvider creates a new ViperProvider
func NewViperProvider(configPath string, configName string, configType string) (*ViperProvider, error) {
	v := viper.New()

	// Set configuration file properties
	v.SetConfigName(configName)
	v.SetConfigType(configType)

	// Add configuration path
	if configPath != "" {
		v.AddConfigPath(configPath)
	} else {
		// Default paths relative to current execution
		v.AddConfigPath(".")            // Current directory
		v.AddConfigPath("config")       // ./config
		v.AddConfigPath("../config")    // ../config (useful for cmd/*)
		v.AddConfigPath("../../config") // ../../config (useful for cmd/subcmd/*)

		// Paths relative to potential project root locations
		v.AddConfigPath("framework/config")       // ./framework/config
		v.AddConfigPath("../framework/config")    // ../framework/config
		v.AddConfigPath("../../framework/config") // ../../framework/config (often works for tests in cmd/*)

		// Absolute paths
		v.AddConfigPath("/etc/app")
	}

	// Set environment variable prefix
	v.SetEnvPrefix("APP")

	// Replace dots with underscores in environment variables
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Enable environment variables
	v.AutomaticEnv()

	// Read configuration file
	if err := v.ReadInConfig(); err != nil {
		// If config file is not found, it's not necessarily an error, might use defaults or env vars
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// Only return error if it's something other than file not found
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	return &ViperProvider{
		viper: v,
	}, nil
}

// Get retrieves a value from the configuration
func (p *ViperProvider) Get(key string) interface{} {
	return p.viper.Get(key)
}

// GetString retrieves a string value from the configuration
func (p *ViperProvider) GetString(key string) string {
	return p.viper.GetString(key)
}

// GetInt retrieves an integer value from the configuration
func (p *ViperProvider) GetInt(key string) int {
	return p.viper.GetInt(key)
}

// GetBool retrieves a boolean value from the configuration
func (p *ViperProvider) GetBool(key string) bool {
	return p.viper.GetBool(key)
}

// GetFloat64 retrieves a float64 value from the configuration
func (p *ViperProvider) GetFloat64(key string) float64 {
	return p.viper.GetFloat64(key)
}

// GetDuration retrieves a duration value from the configuration
func (p *ViperProvider) GetDuration(key string) time.Duration {
	return p.viper.GetDuration(key)
}

// GetStringSlice retrieves a string slice from the configuration
func (p *ViperProvider) GetStringSlice(key string) []string {
	return p.viper.GetStringSlice(key)
}

// GetStringMap retrieves a string map from the configuration
func (p *ViperProvider) GetStringMap(key string) map[string]interface{} {
	return p.viper.GetStringMap(key)
}

// GetStringMapString retrieves a string map of strings from the configuration
func (p *ViperProvider) GetStringMapString(key string) map[string]string {
	return p.viper.GetStringMapString(key)
}

// IsSet checks if a key is set in the configuration
func (p *ViperProvider) IsSet(key string) bool {
	return p.viper.IsSet(key)
}

// AllSettings returns all settings from the configuration
func (p *ViperProvider) AllSettings() map[string]interface{} {
	return p.viper.AllSettings()
}

// LoadConfigFile loads a configuration file
func LoadConfigFile(configPath string) (Provider, error) {
	// If configPath is a directory, use default config name
	if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
		return NewViperProvider(configPath, "config", "yaml")
	}

	// If configPath is a file, extract directory, name, and extension
	dir := filepath.Dir(configPath)
	file := filepath.Base(configPath)
	ext := filepath.Ext(file)
	name := strings.TrimSuffix(file, ext)

	// Remove the dot from extension
	ext = strings.TrimPrefix(ext, ".")

	return NewViperProvider(dir, name, ext)
}

// LoadCLIConfig loads the CLI configuration
func LoadCLIConfig() (Provider, error) {
	// Default CLI config paths
	paths := []string{
		"./framework/config",
		"./config",
		".",
		"../framework/config",    // Added for robustness
		"../../framework/config", // Added for robustness
	}

	for _, path := range paths {
		provider, err := NewViperProvider(path, "cli_config", "yaml")
		// Check if error is nil AND if a key expected in cli_config exists
		if err == nil && provider.IsSet("cli") { // Assuming "cli" is a top-level key
			return provider, nil
		}
	}

	// If no config file found, return a provider based on defaults/env vars
	return NewViperProvider("", "cli_config", "yaml")
}

// LoadServiceConfig loads the service configuration
func LoadServiceConfig() (Provider, error) {
	// Default service config paths
	paths := []string{
		"./framework/config",
		"./config",
		".",
		"../framework/config",    // Added for robustness
		"../../framework/config", // Added for robustness
	}

	for _, path := range paths {
		provider, err := NewViperProvider(path, "service_default", "yaml")
		// Check if error is nil AND if a key expected in service_default exists
		if err == nil && provider.IsSet("app") { // Assuming "app" is a top-level key
			return provider, nil
		}
	}

	// If no config file found, return a provider based on defaults/env vars
	return NewViperProvider("", "service_default", "yaml")
}

// LoadPluginConfig loads the plugin configuration
func LoadPluginConfig() (Provider, error) {
	// Default plugin config paths
	paths := []string{
		"./framework/config",
		"./config",
		".",
		"../framework/config",    // Added for robustness
		"../../framework/config", // Added for robustness
	}

	for _, path := range paths {
		provider, err := NewViperProvider(path, "plugin_settings", "yaml")
		// Check if error is nil AND if a key expected in plugin_settings exists
		if err == nil && provider.IsSet("plugins") { // Assuming "plugins" is a top-level key
			return provider, nil
		}
	}

	// If no config file found, return a provider based on defaults/env vars
	return NewViperProvider("", "plugin_settings", "yaml")
}

// Load loads the application configuration from the specified path or defaults
func Load(configPath string) (*Config, error) {
	var provider Provider
	var err error

	// Determine config file name and path
	configName := "service_default"
	configType := "yaml"
	searchPath := ""

	if configPath != "" {
		// If configPath is a directory, use default config name
		if stat, err := os.Stat(configPath); err == nil && stat.IsDir() {
			searchPath = configPath
		} else if err == nil {
			// If configPath is a file, extract directory, name, and extension
			searchPath = filepath.Dir(configPath)
			file := filepath.Base(configPath)
			ext := filepath.Ext(file)
			configName = strings.TrimSuffix(file, ext)
			configType = strings.TrimPrefix(ext, ".")
		} else {
			// Handle error if configPath is invalid
			return nil, fmt.Errorf("invalid config path %s: %w", configPath, err)
		}
	}

	// Create Viper provider using the determined path/name or default search paths if path is empty
	provider, err = NewViperProvider(searchPath, configName, configType)
	if err != nil {
		return nil, fmt.Errorf("failed to create config provider: %w", err)
	}

	// Unmarshal the config into the Config struct
	var cfg Config
	viperProvider, ok := provider.(*ViperProvider)
	if !ok {
		return nil, fmt.Errorf("unexpected provider type")
	}

	// Set default values (optional, Viper can also handle defaults)
	// viperProvider.viper.SetDefault("app.name", "axiomod-viper-default")
	// viperProvider.viper.SetDefault("app.environment", "development")
	// viperProvider.viper.SetDefault("http.port", 8080)

	if err := viperProvider.viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}
