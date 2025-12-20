package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/axiomod/axiomod/framework/errors"
	"github.com/axiomod/axiomod/platform/observability"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_ctxtags "github.com/grpc-ecosystem/go-grpc-middleware/tags"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

// Server represents a gRPC server
type Server struct {
	server   *grpc.Server
	listener net.Listener
	logger   *observability.Logger
	options  *ServerOptions
}

// ServerOptions contains options for the gRPC server
type ServerOptions struct {
	Host              string
	Port              int
	TLSCertFile       string
	TLSKeyFile        string
	MaxConnectionAge  time.Duration
	MaxConnectionIdle time.Duration
	Timeout           time.Duration
	AuthFunc          grpc_auth.AuthFunc
}

// DefaultServerOptions returns the default server options
func DefaultServerOptions() *ServerOptions {
	return &ServerOptions{
		Host:              "0.0.0.0",
		Port:              9090,
		MaxConnectionAge:  time.Hour,
		MaxConnectionIdle: time.Minute * 15,
		Timeout:           time.Second * 30,
		AuthFunc:          nil,
	}
}

// NewServer creates a new gRPC server
func NewServer(logger *observability.Logger, options *ServerOptions) (*Server, error) {
	if options == nil {
		options = DefaultServerOptions()
	}

	// Create server options
	var serverOptions []grpc.ServerOption

	// Add keepalive parameters
	serverOptions = append(serverOptions, grpc.KeepaliveParams(keepalive.ServerParameters{
		MaxConnectionAge:  options.MaxConnectionAge,
		MaxConnectionIdle: options.MaxConnectionIdle,
	}))

	// Add interceptors
	serverOptions = append(serverOptions, grpc.UnaryInterceptor(
		grpc_middleware.ChainUnaryServer(
			grpc_ctxtags.UnaryServerInterceptor(),
			grpc_zap.UnaryServerInterceptor(logger.Logger),
			grpc_validator.UnaryServerInterceptor(),
			grpc_recovery.UnaryServerInterceptor(
				grpc_recovery.WithRecoveryHandler(recoveryHandler(logger)),
			),
			timeoutInterceptor(options.Timeout),
		),
	))

	// Add auth interceptor if provided
	if options.AuthFunc != nil {
		serverOptions = append(serverOptions, grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_auth.UnaryServerInterceptor(options.AuthFunc),
			),
		))
	}

	// Add TLS if configured
	if options.TLSCertFile != "" && options.TLSKeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(options.TLSCertFile, options.TLSKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
		}
		serverOptions = append(serverOptions, grpc.Creds(creds))
	}

	// Create gRPC server
	server := grpc.NewServer(serverOptions...)

	// Create listener
	addr := fmt.Sprintf("%s:%d", options.Host, options.Port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	// Register health service
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(server, healthServer)

	// Enable reflection
	reflection.Register(server)

	return &Server{
		server:   server,
		listener: listener,
		logger:   logger,
		options:  options,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start() error {
	s.logger.Info("Starting gRPC server", zap.String("address", s.listener.Addr().String()))
	return s.server.Serve(s.listener)
}

// Stop stops the gRPC server
func (s *Server) Stop() {
	s.logger.Info("Stopping gRPC server")
	s.server.GracefulStop()
}

// GetServer returns the underlying gRPC server
func (s *Server) GetServer() *grpc.Server {
	return s.server
}

// RegisterService registers a service with the server
func (s *Server) RegisterService(desc *grpc.ServiceDesc, impl interface{}) {
	s.server.RegisterService(desc, impl)
	s.logger.Info("Registered gRPC service", zap.String("service", desc.ServiceName))
}

// SetServingStatus sets the serving status of a service
func (s *Server) SetServingStatus(service string, status healthpb.HealthCheckResponse_ServingStatus) {
	healthServer := health.NewServer()
	healthServer.SetServingStatus(service, status)
	s.logger.Info("Set gRPC service status", zap.String("service", service), zap.String("status", status.String()))
}

// recoveryHandler handles panics in gRPC handlers
func recoveryHandler(logger *observability.Logger) grpc_recovery.RecoveryHandlerFunc {
	return func(p interface{}) error {
		logger.Error("Recovered from panic in gRPC handler", zap.Any("panic", p))
		return status.Errorf(codes.Internal, "internal server error")
	}
}

// timeoutInterceptor adds a timeout to gRPC requests
func timeoutInterceptor(timeout time.Duration) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		var resp interface{}
		var err error
		done := make(chan struct{})

		go func() {
			resp, err = handler(ctx, req)
			close(done)
		}()

		select {
		case <-done:
			return resp, err
		case <-ctx.Done():
			return nil, errors.Wrap(ctx.Err(), "request timeout")
		}
	}
}
