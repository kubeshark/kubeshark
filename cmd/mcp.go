package cmd

import (
	"github.com/creasty/defaults"
	"github.com/kubeshark/kubeshark/config"
	"github.com/kubeshark/kubeshark/config/configStructs"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var mcpURL string
var mcpKubeconfig string

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

For EKS/GKE clusters that use CLI auth plugins (aws, gcloud), you may also need
to set the PATH environment variable:

  {
    "mcpServers": {
      "kubeshark": {
        "command": "/path/to/kubeshark",
        "args": ["mcp", "--kubeconfig", "/Users/YOUR_USERNAME/.kube/config"],
        "env": {
          "PATH": "/usr/local/bin:/opt/homebrew/bin:/usr/bin:/bin"
        }
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
		// Set kubeconfig path if provided
		if mcpKubeconfig != "" {
			config.Config.Kube.ConfigPathStr = mcpKubeconfig
		}
		setFlags, _ := cmd.Flags().GetStringSlice(config.SetCommandName)
		runMCPWithConfig(setFlags, mcpURL)
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
	mcpCmd.Flags().StringVar(&mcpURL, "url", "", "Direct URL to Kubeshark (e.g., https://kubeshark.example.com). When set, connects directly without kubectl/proxy and disables start/stop/check tools.")
	mcpCmd.Flags().StringVar(&mcpKubeconfig, "kubeconfig", "", "Path to kubeconfig file (e.g., /Users/me/.kube/config)")
}
