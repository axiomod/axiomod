package grpc

import (
	"context"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/platform/observability"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RBACInterceptor is a gRPC interceptor for RBAC enforcement
func RBACInterceptor(rbacService *auth.RBACService, logger *observability.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// In gRPC, we usually get the subject from the context,
		// which should have been populated by an auth interceptor.
		// For this implementation, we'll assume the subject is in the context metadata or similar.
		// This depends on how AuthFunc is implemented.

		// For now, we'll look for a common key in the context.
		// In a real app, you might use a custom context key.
		sub, ok := ctx.Value("username").(string)
		if !ok || sub == "" {
			sub, ok = ctx.Value("user_id").(string)
			if !ok || sub == "" {
				logger.Warn("No subject found in gRPC context")
				return nil, status.Errorf(codes.PermissionDenied, "access denied")
			}
		}

		// Use the full method name as the object, and "call" as the action
		obj := info.FullMethod
		act := "call"

		allowed, err := rbacService.Enforce(sub, obj, act)
		if err != nil {
			logger.Error("RBAC enforcement error in gRPC", zap.Error(err))
			return nil, status.Errorf(codes.Internal, "authorization error")
		}

		if !allowed {
			logger.Warn("gRPC user not authorized",
				zap.String("subject", sub),
				zap.String("object", obj),
				zap.String("action", act),
			)
			return nil, status.Errorf(codes.PermissionDenied, "access denied")
		}

		return handler(ctx, req)
	}
}
