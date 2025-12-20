package integration_test

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLICommandExecution(t *testing.T) {
	// Skip if not in integration test environment
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test; set RUN_INTEGRATION_TESTS=true to run")
	}

	// Test the version command
	cmd := exec.Command("../bin/axiomod", "version")
	output, err := cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Axiomod")

	// Test the help command
	cmd = exec.Command("../bin/axiomod", "--help")
	output, err = cmd.CombinedOutput()
	assert.NoError(t, err)
	assert.Contains(t, string(output), "Usage")
}
