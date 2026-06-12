// Package redis provides a plugin that manages a Redis client connection
// with health checking, for use as a distributed cache or key-value store.
package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/axiomod/axiomod/framework/health"
	"github.com/axiomod/axiomod/framework/observability"

	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const defaultAddr = "localhost:6379"

// Plugin manages a Redis client.
type Plugin struct {
	logger *observability.Logger

	addr     string
	password string
	db       int

	client *goredis.Client
}

// Name returns the name of the plugin.
func (p *Plugin) Name() string {
	return "redis"
}

// Initialize configures the plugin from its settings.
//
// Settings:
//   - addr (string): host:port of the Redis server, default "localhost:6379"
//   - password (string): optional password
//   - db (int): database number, default 0
func (p *Plugin) Initialize(settings map[string]interface{}, logger *observability.Logger, metrics *observability.Metrics, cfg *config.Config, h *health.Health) error {
	p.logger = logger
	p.addr = defaultAddr

	if addr, ok := settings["addr"].(string); ok && addr != "" {
		p.addr = addr
	}
	if password, ok := settings["password"].(string); ok {
		p.password = password
	}
	if db, ok := settings["db"].(int); ok {
		p.db = db
	}

	if h != nil {
		h.RegisterCheck(p.Name(), p.ping)
	}

	return nil
}

// Start connects to Redis and verifies the connection.
func (p *Plugin) Start() error {
	p.client = goredis.NewClient(&goredis.Options{
		Addr:     p.addr,
		Password: p.password,
		DB:       p.db,
	})

	if err := p.ping(); err != nil {
		return fmt.Errorf("redis plugin: failed to connect to %s: %w", p.addr, err)
	}

	if p.logger != nil {
		p.logger.Info("Redis plugin started", zap.String("addr", p.addr), zap.Int("db", p.db))
	}
	return nil
}

// Stop closes the client.
func (p *Plugin) Stop() error {
	if p.client != nil {
		err := p.client.Close()
		p.client = nil
		return err
	}
	return nil
}

// Client returns the underlying Redis client, or nil before Start.
func (p *Plugin) Client() *goredis.Client {
	return p.client
}

// ping verifies the Redis server responds.
func (p *Plugin) ping() error {
	if p.client == nil {
		return fmt.Errorf("redis plugin: not started")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.client.Ping(ctx).Err()
}
