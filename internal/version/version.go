package version

// Version information (can be overridden at build time)
var (
	Version = "0.0.1-alpha"
	Build   = "dev"
	GitHash = "unknown"
)

// BuildInfo returns formatted build information
func BuildInfo() string {
	return "WTF (What's The Function) version " + Version + " (build: " + Build + ", git: " + GitHash + ")"
}
