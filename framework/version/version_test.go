package version

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetInfo(t *testing.T) {
	info := GetInfo()
	assert.Equal(t, Version, info.Version)
	assert.Equal(t, GitCommit, info.GitCommit)
	assert.Equal(t, BuildDate, info.BuildDate)
	assert.NotEmpty(t, info.GoVersion)
	assert.Contains(t, info.Platform, "/")
	assert.False(t, info.Timestamp.IsZero())
}

func TestInfoString(t *testing.T) {
	s := Info{
		Version:   "v1.2.3",
		GitCommit: "abc123",
		BuildDate: "2026-06-12",
		GoVersion: "go1.24",
		Platform:  "linux/amd64",
	}.String()

	for _, fragment := range []string{"v1.2.3", "abc123", "2026-06-12", "go1.24", "linux/amd64"} {
		assert.True(t, strings.Contains(s, fragment), "expected %q in %q", fragment, s)
	}
}

func TestSetVersion(t *testing.T) {
	origVersion, origCommit, origDate := Version, GitCommit, BuildDate
	t.Cleanup(func() { SetVersion(origVersion, origCommit, origDate) })

	SetVersion("v9.9.9", "deadbeef", "2030-01-01")
	info := GetInfo()
	assert.Equal(t, "v9.9.9", info.Version)
	assert.Equal(t, "deadbeef", info.GitCommit)
	assert.Equal(t, "2030-01-01", info.BuildDate)
}
