package auth

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/axiomod/axiomod/framework/config"
	"github.com/stretchr/testify/assert"
)

func TestRBACService(t *testing.T) {
	tempDir := t.TempDir()

	modelContent := `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
`
	policyContent := `
p, alice, data1, read
p, bob, data2, write
`
	modelPath := filepath.Join(tempDir, "model.conf")
	policyPath := filepath.Join(tempDir, "policy.csv")

	err := os.WriteFile(modelPath, []byte(modelContent), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(policyPath, []byte(policyContent), 0644)
	assert.NoError(t, err)

	cfg := config.CasbinConfig{
		ModelPath:  modelPath,
		PolicyPath: policyPath,
	}

	service, err := NewRBACService(cfg)
	assert.NoError(t, err)

	t.Run("Enforce Allowed", func(t *testing.T) {
		allowed, err := service.Enforce("alice", "data1", "read")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})

	t.Run("Enforce Denied", func(t *testing.T) {
		allowed, err := service.Enforce("alice", "data2", "write")
		assert.NoError(t, err)
		assert.False(t, allowed)
	})

	t.Run("Reload Policy", func(t *testing.T) {
		// Add new policy
		newPolicyContent := policyContent + "p, jane, data3, read\n"
		err := os.WriteFile(policyPath, []byte(newPolicyContent), 0644)
		assert.NoError(t, err)

		err = service.ReloadPolicy()
		assert.NoError(t, err)

		allowed, err := service.Enforce("jane", "data3", "read")
		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}
