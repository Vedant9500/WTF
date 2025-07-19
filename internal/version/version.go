package version

// Version information (can be overridden at build time)
var (
	Version = "1.1.0"
	Build   = "release"
	GitHash = "unknown"
)

// BuildInfo returns formatted build information
func BuildInfo() string {
	return "WTF (What's The Function) version " + Version + " (build: " + Build + ", git: " + GitHash + ")"
}
