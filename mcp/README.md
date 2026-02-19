# Kubeshark MCP Server

[Kubeshark](https://kubeshark.com) MCP (Model Context Protocol) server enables AI assistants like Claude Desktop, Cursor, and other MCP-compatible clients to query real-time Kubernetes network traffic.

## Features

- **L7 API Traffic Analysis**: Query HTTP, gRPC, Redis, Kafka, DNS transactions
- **L4 Network Flows**: View TCP/UDP flows with traffic statistics
- **Cluster Management**: Start/stop Kubeshark deployments (with safety controls)
- **PCAP Snapshots**: Create and export network captures
- **Built-in Prompts**: Pre-configured prompts for common analysis tasks

## Installation

### 1. Install Kubeshark CLI

```bash
# macOS
brew install kubeshark

# Linux
sh <(curl -Ls https://kubeshark.com/install)

# Windows (PowerShell)
choco install kubeshark
```

Or download from [GitHub Releases](https://github.com/kubeshark/kubeshark/releases).

### 2. Configure Claude Desktop

Add to your Claude Desktop configuration:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
**Windows**: `%APPDATA%\Claude\claude_desktop_config.json`

#### URL Mode (Recommended for existing deployments)

```json
{
  "mcpServers": {
    "kubeshark": {
      "command": "kubeshark",
      "args": ["mcp", "--url", "https://kubeshark.example.com"]
    }
  }
}
```

#### Proxy Mode (Requires kubectl access)

```json
{
  "mcpServers": {
    "kubeshark": {
      "command": "kubeshark",
      "args": ["mcp", "--kubeconfig", "/path/to/.kube/config"]
    }
  }
}
```
or:

```json
{
  "mcpServers": {
    "kubeshark": {
      "command": "kubeshark",
      "args": ["mcp"]
    }
  }
}
```

#### With Destructive Operations

```json
{
  "mcpServers": {
    "kubeshark": {
      "command": "kubeshark",
      "args": ["mcp", "--allow-destructive", "--kubeconfig", "/path/to/.kube/config"]
    }
  }
}
```

### 3. Generate Configuration

Use the CLI to generate configuration:

```bash
kubeshark mcp --mcp-config --url https://kubeshark.example.com
```

## Available Tools

### Traffic Analysis (All Modes)

| Tool | Description |
|------|-------------|
| `list_workloads` | List pods, services, namespaces with observed traffic |
| `list_api_calls` | Query L7 API transactions with KFL filtering |
| `get_api_call` | Get detailed info about a specific API call |
| `get_api_stats` | Get aggregated API statistics |
| `list_l4_flows` | List L4 (TCP/UDP) network flows |
| `get_l4_flow_summary` | Get L4 connectivity summary |
| `list_snapshots` | List all PCAP snapshots |
| `create_snapshot` | Create a new PCAP snapshot |
| `get_dissection_status` | Check L7 protocol parsing status |
| `enable_dissection` | Enable L7 protocol dissection |
| `disable_dissection` | Disable L7 protocol dissection |

### Cluster Management (Proxy Mode Only)

| Tool | Description | Requires |
|------|-------------|----------|
| `check_kubeshark_status` | Check if Kubeshark is running | - |
| `start_kubeshark` | Deploy Kubeshark to cluster | `--allow-destructive` |
| `stop_kubeshark` | Remove Kubeshark from cluster | `--allow-destructive` |

## Available Prompts

| Prompt | Description |
|--------|-------------|
| `analyze_traffic` | Analyze API traffic patterns and identify issues |
| `find_errors` | Find and summarize API errors and failures |
| `trace_request` | Trace a request path through microservices |
| `show_topology` | Show service communication topology |
| `latency_analysis` | Analyze latency patterns and identify slow endpoints |
| `security_audit` | Audit traffic for security concerns |
| `compare_traffic` | Compare traffic patterns between time periods |
| `debug_connection` | Debug connectivity issues between services |

## Example Conversations

```
User: Show me all HTTP 500 errors in the last hour

Claude: I'll query the API traffic for 500 errors.
[Calling list_api_calls with kfl="http and response.status == 500"]

Found 12 HTTP 500 errors:
1. POST /api/checkout -> payment-service (500)
   Time: 10:23:45 | Latency: 2340ms
...
```

```
User: What services are communicating with the database?

Claude: Let me check the L4 flows to the database.
[Calling list_l4_flows with dst_filter="postgres"]

Found 5 services connecting to postgres:5432:
- orders-service: 456KB transferred
- users-service: 123KB transferred
...
```

## CLI Options

| Option | Description |
|--------|-------------|
| `--url` | Direct URL to Kubeshark Hub |
| `--kubeconfig` | Path to kubeconfig file |
| `--allow-destructive` | Enable start/stop operations |
| `--list-tools` | List available tools and exit |
| `--mcp-config` | Print Claude Desktop config JSON |

## KFL (Kubeshark Filter Language)

Query traffic using KFL syntax:

```
# HTTP requests to a specific path
http and request.path == "/api/users"

# Errors only
response.status >= 400

# Specific source pod
src.pod.name == "frontend-.*"

# Multiple conditions
http and src.namespace == "default" and response.status == 500
```

## MCP Registry

Kubeshark is published to the [MCP Registry](https://registry.modelcontextprotocol.io/) automatically on each release.

The `server.json` in this directory is a reference file. The actual registry metadata (version, SHA256 hashes) is auto-generated during the release workflow. See [`.github/workflows/release.yml`](../.github/workflows/release.yml) for details.

## Links

- [Documentation](https://docs.kubeshark.com/en/mcp)
- [GitHub](https://github.com/kubeshark/kubeshark)
- [Website](https://kubeshark.com)
- [MCP Registry](https://registry.modelcontextprotocol.io/)

## License

Apache-2.0
