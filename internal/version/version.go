// Package version provides build-time version information for the WTF application.
//
// Version information can be set at build time using ldflags:
//
//	go build -ldflags "-X github.com/Vedant9500/WTF/internal/version.Version=1.2.3"
package version

// Version information (can be overridden at build time)
var (
	// Version is the semantic version of the application
	Version = "1.2.0"

	// Build indicates the build type (release, debug, etc.)
	Build = "release"

	// GitHash contains the git commit hash at build time
	GitHash = "unknown"
)

// BuildInfo returns formatted build information
func BuildInfo() string {
	return "WTF (What's The Function) version " + Version + " (build: " + Build + ", git: " + GitHash + ")"
}
