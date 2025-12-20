package version

import (
	"fmt"
	"runtime"
	"time"
)

var (
	// Version is the version of the application
	Version = "dev"

	// GitCommit is the git commit hash
	GitCommit = "unknown"

	// BuildDate is the date when the application was built
	BuildDate = "unknown"

	// GoVersion is the Go version used to build the application
	GoVersion = runtime.Version()

	// Platform is the platform the application was built for
	Platform = fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
)

// Info contains version information
type Info struct {
	Version   string    `json:"version"`
	GitCommit string    `json:"git_commit"`
	BuildDate string    `json:"build_date"`
	GoVersion string    `json:"go_version"`
	Platform  string    `json:"platform"`
	Timestamp time.Time `json:"timestamp"`
}

// GetInfo returns version information
func GetInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  Platform,
		Timestamp: time.Now(),
	}
}

// String returns a string representation of version information
func (i Info) String() string {
	return fmt.Sprintf(
		"Version: %s\nGit Commit: %s\nBuild Date: %s\nGo Version: %s\nPlatform: %s",
		i.Version,
		i.GitCommit,
		i.BuildDate,
		i.GoVersion,
		i.Platform,
	)
}

// SetVersion sets the version information
func SetVersion(version, gitCommit, buildDate string) {
	Version = version
	GitCommit = gitCommit
	BuildDate = buildDate
}

// BuildInfoFromLDFlags sets version information from ldflags
// Usage: go build -ldflags "-X axiomod/internal/framework/version.Version=1.0.0 -X axiomod/internal/framework/version.GitCommit=$(git rev-parse HEAD) -X axiomod/internal/framework/version.BuildDate=$(date -u +'%Y-%m-%dT%H:%M:%SZ')"
func BuildInfoFromLDFlags() {
	// This function doesn't need to do anything, it's just a placeholder for ldflags
}
