package grpc

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const testCasbinModel = `[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

const testCasbinPolicy = `p, alice, /pkg.Service/Method, call
`

// newTestRBACService creates an RBAC service backed by temp model/policy files.
func newTestRBACService(t *testing.T) *auth.RBACService {
	t.Helper()
	dir := t.TempDir()

	modelPath := filepath.Join(dir, "model.conf")
	require.NoError(t, os.WriteFile(modelPath, []byte(testCasbinModel), 0644))

	policyPath := filepath.Join(dir, "policy.csv")
	require.NoError(t, os.WriteFile(policyPath, []byte(testCasbinPolicy), 0644))

	service, err := auth.NewRBACService(config.CasbinConfig{
		ModelPath:  modelPath,
		PolicyPath: policyPath,
	})
	require.NoError(t, err)
	return service
}

func TestRBACInterceptor_NoSubjectDenied(t *testing.T) {
	interceptor := RBACInterceptor(newTestRBACService(t), testLogger())
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}

	handlerCalled := false
	resp, err := interceptor(context.Background(), "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return "response", nil
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.False(t, handlerCalled)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestRBACInterceptor_AllowedSubject(t *testing.T) {
	interceptor := RBACInterceptor(newTestRBACService(t), testLogger())
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}

	//nolint:staticcheck // The interceptor reads plain string context keys.
	ctx := context.WithValue(context.Background(), "username", "alice")

	resp, err := interceptor(ctx, "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
}

func TestRBACInterceptor_DeniedSubject(t *testing.T) {
	interceptor := RBACInterceptor(newTestRBACService(t), testLogger())
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}

	//nolint:staticcheck // The interceptor reads plain string context keys.
	ctx := context.WithValue(context.Background(), "username", "mallory")

	resp, err := interceptor(ctx, "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	})

	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, codes.PermissionDenied, status.Code(err))
}

func TestRBACInterceptor_UserIDFallback(t *testing.T) {
	interceptor := RBACInterceptor(newTestRBACService(t), testLogger())
	info := &grpc.UnaryServerInfo{FullMethod: "/pkg.Service/Method"}

	//nolint:staticcheck // The interceptor reads plain string context keys.
	ctx := context.WithValue(context.Background(), "user_id", "alice")

	resp, err := interceptor(ctx, "request", info, func(ctx context.Context, req interface{}) (interface{}, error) {
		return "response", nil
	})

	require.NoError(t, err)
	assert.Equal(t, "response", resp)
}
