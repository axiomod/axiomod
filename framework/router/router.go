package router

import (
	"github.com/axiomod/axiomod/platform/observability"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/favicon"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"go.uber.org/zap"
)

// Config contains configuration for the router
type Config struct {
	// Prefork determines whether to use the prefork feature
	Prefork bool
	// ServerHeader is the server header
	ServerHeader string
	// StrictRouting determines whether routes are case-sensitive
	StrictRouting bool
	// CaseSensitive determines whether paths are case-sensitive
	CaseSensitive bool
	// BodyLimit is the maximum allowed size for a request body
	BodyLimit int
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout int
	// WriteTimeout is the maximum duration for writing the response
	WriteTimeout int
	// IdleTimeout is the maximum amount of time to wait for the next request
	IdleTimeout int
	// EnableCORS determines whether to enable CORS
	EnableCORS bool
	// EnableCompression determines whether to enable compression
	EnableCompression bool
	// EnableETag determines whether to enable ETag
	EnableETag bool
	// EnableFavicon determines whether to enable favicon
	EnableFavicon bool
	// EnableLimiter determines whether to enable rate limiting
	EnableLimiter bool
	// EnableRecover determines whether to enable panic recovery
	EnableRecover bool
	// EnableRequestID determines whether to enable request ID
	EnableRequestID bool
}

// DefaultConfig returns the default router configuration
func DefaultConfig() *Config {
	return &Config{
		Prefork:           false,
		ServerHeader:      "Go-Macroservice",
		StrictRouting:     false,
		CaseSensitive:     false,
		BodyLimit:         4 * 1024 * 1024, // 4MB
		ReadTimeout:       60,
		WriteTimeout:      60,
		IdleTimeout:       120,
		EnableCORS:        true,
		EnableCompression: true,
		EnableETag:        true,
		EnableFavicon:     true,
		EnableLimiter:     false,
		EnableRecover:     true,
		EnableRequestID:   true,
	}
}

// Router is a wrapper around fiber.App
type Router struct {
	app    *fiber.App
	logger *observability.Logger
	config *Config
}

// New creates a new router
func New(logger *observability.Logger, config *Config) *Router {
	if config == nil {
		config = DefaultConfig()
	}

	// Create fiber app
	app := fiber.New(fiber.Config{
		Prefork:       config.Prefork,
		ServerHeader:  config.ServerHeader,
		StrictRouting: config.StrictRouting,
		CaseSensitive: config.CaseSensitive,
		BodyLimit:     config.BodyLimit,
		ReadTimeout:   time.Duration(config.ReadTimeout) * time.Second,
		WriteTimeout:  time.Duration(config.WriteTimeout) * time.Second,
		IdleTimeout:   time.Duration(config.IdleTimeout) * time.Second,
	})

	// Add middleware
	if config.EnableCORS {
		app.Use(cors.New())
	}

	if config.EnableCompression {
		app.Use(compress.New())
	}

	if config.EnableETag {
		app.Use(etag.New())
	}

	if config.EnableFavicon {
		app.Use(favicon.New())
	}

	if config.EnableLimiter {
		app.Use(limiter.New())
	}

	if config.EnableRecover {
		app.Use(recover.New())
	}

	if config.EnableRequestID {
		app.Use(requestid.New())
	}

	logger.Info("Created router", zap.Bool("prefork", config.Prefork))

	return &Router{
		app:    app,
		logger: logger,
		config: config,
	}
}

// App returns the underlying fiber.App
func (r *Router) App() *fiber.App {
	return r.app
}

// Group creates a new router group
func (r *Router) Group(prefix string, handlers ...fiber.Handler) fiber.Router {
	return r.app.Group(prefix, handlers...)
}

// Get registers a route for GET method
func (r *Router) Get(path string, handler fiber.Handler) fiber.Router {
	return r.app.Get(path, handler)
}

// Post registers a route for POST method
func (r *Router) Post(path string, handler fiber.Handler) fiber.Router {
	return r.app.Post(path, handler)
}

// Put registers a route for PUT method
func (r *Router) Put(path string, handler fiber.Handler) fiber.Router {
	return r.app.Put(path, handler)
}

// Delete registers a route for DELETE method
func (r *Router) Delete(path string, handler fiber.Handler) fiber.Router {
	return r.app.Delete(path, handler)
}

// Patch registers a route for PATCH method
func (r *Router) Patch(path string, handler fiber.Handler) fiber.Router {
	return r.app.Patch(path, handler)
}

// Options registers a route for OPTIONS method
func (r *Router) Options(path string, handler fiber.Handler) fiber.Router {
	return r.app.Options(path, handler)
}

// Head registers a route for HEAD method
func (r *Router) Head(path string, handler fiber.Handler) fiber.Router {
	return r.app.Head(path, handler)
}

// All registers a route for all HTTP methods
func (r *Router) All(path string, handler fiber.Handler) fiber.Router {
	return r.app.All(path, handler)
}

// Use registers middleware
func (r *Router) Use(args ...interface{}) fiber.Router {
	return r.app.Use(args...)
}

// Static serves static files
func (r *Router) Static(prefix, root string) fiber.Router {
	return r.app.Static(prefix, root)
}

// Listen starts the HTTP server
func (r *Router) Listen(addr string) error {
	r.logger.Info("Starting HTTP server", zap.String("address", addr))
	return r.app.Listen(addr)
}

// Shutdown gracefully shuts down the server
func (r *Router) Shutdown() error {
	r.logger.Info("Shutting down HTTP server")
	return r.app.Shutdown()
}
