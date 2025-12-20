package config

// Config represents the application configuration
type Config struct {
	App           AppConfig
	Observability ObservabilityConfig
	Database      DatabaseConfig
	HTTP          HTTPConfig
	GRPC          GRPCConfig
	Plugins       PluginsConfig
}

// AppConfig represents the application-specific configuration
type AppConfig struct {
	Name        string
	Environment string
	Version     string
	Debug       bool
}

// ObservabilityConfig represents the observability configuration
type ObservabilityConfig struct {
	LogLevel       string
	LogFormat      string
	TracingEnabled bool
	TracingURL     string
	MetricsEnabled bool
	MetricsPort    int
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Driver   string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

// HTTPConfig represents the HTTP server configuration
type HTTPConfig struct {
	Port         int
	Host         string
	ReadTimeout  int
	WriteTimeout int
}

// GRPCConfig represents the gRPC server configuration
type GRPCConfig struct {
	Port int
	Host string
}
