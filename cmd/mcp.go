package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mcpTapSetFlags []string
var mcpURL string

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Run MCP (Model Context Protocol) server for AI assistant integration",
	Long: `Run an MCP server over stdio that exposes Kubeshark's L7 API visibility
to AI assistants like Claude Desktop.

TOOLS PROVIDED:

Cluster Management (work without Kubeshark running):
  - check_kubeshark_status: Check if Kubeshark is running in the cluster
  - start_kubeshark: Start Kubeshark to capture traffic
  - stop_kubeshark: Stop Kubeshark and clean up resources

Traffic Analysis (require Kubeshark running):
  - list_workloads: Discover pods, services, namespaces, and nodes with L7 traffic
  - list_api_calls: Query L7 API transactions (HTTP, gRPC, etc.)
  - get_api_call: Get detailed information about a specific API call
  - get_api_stats: Get aggregated API statistics

CONFIGURATION:

To use with Claude Desktop, add to your claude_desktop_config.json
(typically at ~/Library/Application Support/Claude/claude_desktop_config.json):

  {
    "mcpServers": {
      "kubeshark": {
        "command": "/path/to/kubeshark",
        "args": ["mcp"],
        "env": {
          "PATH": "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin",
          "HOME": "/Users/YOUR_USERNAME",
          "KUBECONFIG": "/Users/YOUR_USERNAME/.kube/config"
        }
      }
    }
  }

IMPORTANT: The "env" section is required because MCP servers run in a sandboxed
environment without access to your shell's PATH or environment variables. Without
it, kubectl commands will fail with authentication errors.

For EKS clusters, ensure /usr/local/bin is in PATH (for aws CLI).
For GKE clusters, ensure gcloud is accessible in PATH.

DIRECT URL MODE:

If Kubeshark is already running and accessible via URL (e.g., exposed via ingress),
you can connect directly without needing kubectl/kubeconfig:

  {
    "mcpServers": {
      "kubeshark": {
        "command": "/path/to/kubeshark",
        "args": ["mcp", "--url", "https://kubeshark.example.com"]
      }
    }
  }

In URL mode, cluster management tools (start/stop/check) are disabled since
Kubeshark is managed externally.

CUSTOM DOCKER IMAGES:

To use custom Docker images when starting Kubeshark, add --tap-set flags:

  {
    "mcpServers": {
      "kubeshark": {
        "command": "/path/to/kubeshark",
        "args": ["mcp", "--tap-set", "tap.docker.tag=v52.3"],
        ...
      }
    }
  }

Multiple --tap-set flags can be used for different settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		runMCPWithConfig(mcpTapSetFlags, mcpURL)
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
	mcpCmd.Flags().StringArrayVar(&mcpTapSetFlags, "tap-set", []string{}, "Set values to pass to 'kubeshark tap' when using start_kubeshark (can be used multiple times)")
	mcpCmd.Flags().StringVar(&mcpURL, "url", "", "Direct URL to Kubeshark (e.g., https://kubeshark.example.com). When set, connects directly without kubectl/proxy and disables start/stop/check tools.")
}
