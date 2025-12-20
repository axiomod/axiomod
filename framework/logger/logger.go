package logger

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config contains configuration for the logger
type Config struct {
	// Level is the minimum enabled logging level
	Level string
	// Format is the log format (json or console)
	Format string
	// OutputPaths is a list of URLs or file paths to write logging output to
	OutputPaths []string
	// ErrorOutputPaths is a list of URLs or file paths to write internal logger errors to
	ErrorOutputPaths []string
	// Development puts the logger in development mode
	Development bool
	// DisableCaller stops annotating logs with the calling function's file name and line number
	DisableCaller bool
	// DisableStacktrace disables automatic stacktrace capturing
	DisableStacktrace bool
	// Sampling configures sampling of logs
	Sampling *SamplingConfig
}

// SamplingConfig contains configuration for log sampling
type SamplingConfig struct {
	// Initial is the initial sampling rate (first N entries per second)
	Initial int
	// Thereafter is the sampling rate after the initial rate is exceeded
	Thereafter int
}

// DefaultConfig returns the default logger configuration
func DefaultConfig() *Config {
	return &Config{
		Level:             "info",
		Format:            "json",
		OutputPaths:       []string{"stdout"},
		ErrorOutputPaths:  []string{"stderr"},
		Development:       false,
		DisableCaller:     false,
		DisableStacktrace: false,
		Sampling: &SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
	}
}

// Logger is a wrapper around zap.Logger
type Logger struct {
	*zap.Logger
	config *Config
}

// New creates a new logger with the given configuration
func New(config *Config) (*Logger, error) {
	if config == nil {
		config = DefaultConfig()
	}

	// Parse log level
	level := zap.InfoLevel
	if err := level.UnmarshalText([]byte(config.Level)); err != nil {
		return nil, err
	}

	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create zap config
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       config.Development,
		DisableCaller:     config.DisableCaller,
		DisableStacktrace: config.DisableStacktrace,
		Sampling: &zap.SamplingConfig{
			Initial:    config.Sampling.Initial,
			Thereafter: config.Sampling.Thereafter,
		},
		Encoding:         config.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      config.OutputPaths,
		ErrorOutputPaths: config.ErrorOutputPaths,
	}

	// Build logger
	logger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	return &Logger{
		Logger: logger,
		config: config,
	}, nil
}

// NewDevelopment creates a new development logger
func NewDevelopment() (*Logger, error) {
	config := DefaultConfig()
	config.Development = true
	config.Format = "console"
	config.Level = "debug"
	config.DisableStacktrace = true
	return New(config)
}

// NewProduction creates a new production logger
func NewProduction() (*Logger, error) {
	config := DefaultConfig()
	config.Development = false
	config.Format = "json"
	config.Level = "info"
	return New(config)
}

// With creates a child logger with the given fields
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
	}
}

// Named creates a child logger with the given name
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger: l.Logger.Named(name),
		config: l.config,
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// StdLogger returns a standard library logger that forwards to the zap logger
func (l *Logger) StdLogger() *zap.SugaredLogger {
	return l.Logger.Sugar()
}

// SetLevel sets the logging level
func (l *Logger) SetLevel(level string) error {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		return err
	}

	l.config.Level = level
	return nil
}

// GetLevel returns the current logging level
func (l *Logger) GetLevel() string {
	return l.config.Level
}

// GetConfig returns the logger configuration
func (l *Logger) GetConfig() *Config {
	return l.config
}

// NewTestLogger creates a logger for testing
func NewTestLogger() *Logger {
	config := DefaultConfig()
	config.Development = true
	config.Format = "console"
	config.Level = "debug"
	config.DisableStacktrace = true
	config.OutputPaths = []string{os.DevNull}
	config.ErrorOutputPaths = []string{os.DevNull}

	logger, _ := New(config)
	return logger
}

// Fields creates a set of fields from key-value pairs
func Fields(keysAndValues ...interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key, ok := keysAndValues[i].(string)
			if !ok {
				key = "unknown"
			}
			fields = append(fields, zap.Any(key, keysAndValues[i+1]))
		}
	}
	return fields
}

// WithContext returns a logger with context fields
func (l *Logger) WithContext(ctx map[string]interface{}) *Logger {
	if len(ctx) == 0 {
		return l
	}

	fields := make([]zap.Field, 0, len(ctx))
	for k, v := range ctx {
		fields = append(fields, zap.Any(k, v))
	}

	return &Logger{
		Logger: l.Logger.With(fields...),
		config: l.config,
	}
}

// WithTimestamp returns a logger with a timestamp field
func (l *Logger) WithTimestamp() *Logger {
	return &Logger{
		Logger: l.Logger.With(zap.Time("timestamp", time.Now())),
		config: l.config,
	}
}
