package cmd

import (
	"github.com/kubeshark/kubeshark/config"
	"github.com/spf13/cobra"
)

var mcpURL string
var mcpKubeconfig string
var mcpListTools bool
var mcpConfig bool
var mcpAllowDestructive bool

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
        "args": ["mcp", "--kubeconfig", "/Users/YOUR_USERNAME/.kube/config"]
      }
    }
  }

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

DESTRUCTIVE OPERATIONS:

By default, destructive operations (start_kubeshark, stop_kubeshark) are disabled
to prevent accidental cluster modifications. To enable them, use --allow-destructive:

  {
    "mcpServers": {
      "kubeshark": {
        "command": "/path/to/kubeshark",
        "args": ["mcp", "--allow-destructive", "--kubeconfig", "/path/to/.kube/config"]
      }
    }
  }

CUSTOM SETTINGS:

To use custom settings when starting Kubeshark, use the --set flag:

  {
    "mcpServers": {
      "kubeshark": {
        "command": "/path/to/kubeshark",
        "args": ["mcp", "--set", "tap.docker.tag=v52.3"],
        ...
      }
    }
  }

Multiple --set flags can be used for different settings.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle --mcp-config flag
		if mcpConfig {
			printMCPConfig(mcpURL, mcpKubeconfig)
			return nil
		}

		// Set kubeconfig path if provided
		if mcpKubeconfig != "" {
			config.Config.Kube.ConfigPathStr = mcpKubeconfig
		}

		// Handle --list-tools flag
		if mcpListTools {
			listMCPTools(mcpURL)
			return nil
		}

		setFlags, _ := cmd.Flags().GetStringSlice(config.SetCommandName)
		runMCPWithConfig(setFlags, mcpURL, mcpAllowDestructive)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)

	mcpCmd.Flags().StringVar(&mcpURL, "url", "", "Direct URL to Kubeshark (e.g., https://kubeshark.example.com). When set, connects directly without kubectl/proxy and disables start/stop/check tools.")
	mcpCmd.Flags().StringVar(&mcpKubeconfig, "kubeconfig", "", "Path to kubeconfig file (e.g., /Users/me/.kube/config)")
	mcpCmd.Flags().BoolVar(&mcpListTools, "list-tools", false, "List available MCP tools and exit")
	mcpCmd.Flags().BoolVar(&mcpConfig, "mcp-config", false, "Print MCP client configuration JSON and exit")
	mcpCmd.Flags().BoolVar(&mcpAllowDestructive, "allow-destructive", false, "Enable destructive operations (start_kubeshark, stop_kubeshark). Without this flag, only read-only traffic analysis tools are available.")
}
