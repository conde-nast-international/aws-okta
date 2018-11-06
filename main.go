package main

import (
	"github.com/conde-nast-international/aws-okta/cmd"
)

// These are set via linker flags
var (
	Version           = "v0.20.6-conde"
	AnalyticsWriteKey = ""
)

func main() {
	// vars set by linker flags must be strings...
	cmd.Execute(Version, AnalyticsWriteKey)
}
