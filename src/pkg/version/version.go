package version

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

// These variables are set during build time using ldflags
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
	GoVersion = runtime.Version()
)

// Info holds version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

// Get returns the version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: GoVersion,
	}
}

// String returns a formatted version string
func String() string {
	info := Get()

	// Format build time if it's not "unknown"
	var buildTimeStr string
	if info.BuildTime != "unknown" {
		if t, err := time.Parse(time.RFC3339, info.BuildTime); err == nil {
			buildTimeStr = t.Format("2006-01-02 15:04:05")
		} else {
			buildTimeStr = info.BuildTime
		}
	} else {
		buildTimeStr = "unknown"
	}

	// Short git commit (first 7 characters)
	gitCommitShort := info.GitCommit
	if len(gitCommitShort) > 7 {
		gitCommitShort = gitCommitShort[:7]
	}

	return fmt.Sprintf("Weave CLI %s\n"+
		"  Git Commit: %s\n"+
		"  Build Time: %s\n"+
		"  Go Version: %s",
		info.Version,
		gitCommitShort,
		buildTimeStr,
		info.GoVersion)
}

// Short returns a short version string
func Short() string {
	return fmt.Sprintf("Weave CLI %s", Version)
}

// IsRelease returns true if this is a release version (not dev)
func IsRelease() bool {
	return !strings.HasPrefix(Version, "dev") && Version != "unknown"
}

// GetVersionFromGitTag extracts version from git tag
func GetVersionFromGitTag() string {
	// This will be set by the build script
	return Version
}
