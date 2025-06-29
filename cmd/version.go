package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"
)

// getVersion returns version information from build info or fallback
func getVersion() (version, commit, date string) {
	version = "dev"
	commit = "unknown"
	date = "unknown"
	
	if info, ok := debug.ReadBuildInfo(); ok {
		// Get version from module info
		if info.Main.Version != "(devel)" && info.Main.Version != "" {
			version = info.Main.Version
		}
		
		// Extract build settings
		for _, setting := range info.Settings {
			switch setting.Key {
			case "vcs.revision":
				if len(setting.Value) >= 7 {
					commit = setting.Value[:7] // Short hash
				}
			case "vcs.time":
				if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
					date = t.Format("2006-01-02T15:04:05Z")
				}
			case "vcs.modified":
				if setting.Value == "true" {
					version += "-dirty"
				}
			}
		}
	}
	
	return version, commit, date
}

// BuildInfo returns formatted build information
func BuildInfo() string {
	version, commit, date := getVersion()
	return fmt.Sprintf("gocamelpack %s\nCommit: %s\nBuilt: %s\nGo: %s", 
		version, commit, date, runtime.Version())
}

// Version returns just the version string
func Version() string {
	version, _, _ := getVersion()
	return version
}