package ldap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlugin_Lifecycle(t *testing.T) {
	p := &Plugin{}

	// Test Name
	assert.Equal(t, "ldap", p.Name())

	// Test Initialize
	err := p.Initialize(map[string]interface{}{}, nil, nil, nil, nil)
	assert.NoError(t, err)

	// Test Start
	err = p.Start()
	assert.NoError(t, err)

	// Test Stop
	err = p.Stop()
	assert.NoError(t, err)
}
