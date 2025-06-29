package cmd

import (
	"fmt"
	"runtime"
)

// These variables are set via build flags
var (
	Version   = "dev"
	CommitSHA = "unknown"
	BuildDate = "unknown"
)

// BuildInfo returns formatted build information
func BuildInfo() string {
	return fmt.Sprintf("gocamelpack %s\nCommit: %s\nBuilt: %s\nGo: %s", 
		Version, CommitSHA, BuildDate, runtime.Version())
}