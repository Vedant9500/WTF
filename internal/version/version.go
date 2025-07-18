package version

// Version information (can be overridden at build time)
var (
	Version = "1.0.0-dev"
	Build   = "dev"
	GitHash = "unknown"
)

// BuildInfo returns formatted build information
func BuildInfo() string {
	return "cmd-finder version " + Version + " (build: " + Build + ", git: " + GitHash + ")"
}
