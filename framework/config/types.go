package config

// Config represents the application configuration
type Config struct {
	App           AppConfig
	Observability ObservabilityConfig
	Database      DatabaseConfig
	HTTP          HTTPConfig
	GRPC          GRPCConfig
	Auth          AuthConfig
	Casbin        CasbinConfig
	Plugins       PluginsConfig
}

// AuthConfig represents the authentication configuration
type AuthConfig struct {
	OIDC OIDCConfig
	JWT  JWTConfig
}

// OIDCConfig represents the OIDC configuration
type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	JWKSCacheTTL int // in minutes
}

// JWTConfig represents the JWT configuration
type JWTConfig struct {
	SecretKey     string
	TokenDuration int // in minutes
}

// CasbinConfig represents the Casbin RBAC configuration
type CasbinConfig struct {
	ModelPath  string
	PolicyPath string
	Table      string // for database policy
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
	LogLevel            string
	LogFormat           string
	TracingEnabled      bool
	TracingExporterType string // "otlp", "jaeger", "stdout"
	TracingURL          string
	TracingSamplerRatio float64
	MetricsEnabled      bool
	MetricsPort         int
}

// DatabaseConfig represents the database configuration
type DatabaseConfig struct {
	Driver             string
	Host               string
	Port               int
	User               string
	Password           string
	Name               string
	SSLMode            string
	MaxOpenConns       int
	MaxIdleConns       int
	ConnMaxLifetime    int // in minutes
	SlowQueryThreshold int // in milliseconds
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
