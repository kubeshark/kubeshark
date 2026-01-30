package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run MCP (Model Context Protocol) server for AI assistant integration",
	Long: `Run an MCP server over stdio that exposes Kubeshark's L7 API visibility
to AI assistants like Claude Desktop.

The MCP server establishes a connection to Kubeshark via port-forward and
provides tools for:
  - list_workloads: Discover pods, services, namespaces, and nodes with L7 traffic
  - list_api_calls: Query L7 API transactions (HTTP, gRPC, etc.)
  - get_api_call: Get detailed information about a specific API call
  - get_api_stats: Get aggregated API statistics

To use with Claude Desktop, add to your claude_desktop_config.json:
  {
    "mcpServers": {
      "kubeshark": {
        "command": "kubeshark",
        "args": ["mcp"]
      }
    }
  }`,
	RunE: func(cmd *cobra.Command, args []string) error {
		runMCP()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	defaultTapConfig := configStructs.TapConfig{}
	if err := defaults.Set(&defaultTapConfig); err != nil {
		log.Debug().Err(err).Send()
	}

	mcpCmd.Flags().Uint16(configStructs.ProxyFrontPortLabel, defaultTapConfig.Proxy.Front.Port, "Provide a custom port for the proxy/port-forward")
	mcpCmd.Flags().String(configStructs.ProxyHostLabel, defaultTapConfig.Proxy.Host, "Provide a custom host for the proxy/port-forward")
	mcpCmd.Flags().StringP(configStructs.ReleaseNamespaceLabel, "s", defaultTapConfig.Release.Namespace, "Release namespace of Kubeshark")
}
