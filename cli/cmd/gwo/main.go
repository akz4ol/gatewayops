// GatewayOps CLI - Command line interface for the GatewayOps MCP Gateway
package main

import (
	"os"

	"github.com/akz4ol/gatewayops/cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
