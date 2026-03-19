# Kubeshark MCP Setup Reference

## Installing the CLI

**Homebrew (macOS)**:
```bash
brew install kubeshark
```

**Linux**:
```bash
sh <(curl -Ls https://kubeshark.com/install)
```

**From source**:
```bash
git clone https://github.com/kubeshark/kubeshark
cd kubeshark && make
```

## MCP Configuration

**Claude Desktop / Cowork** (`claude_desktop_config.json`):
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

**Claude Code (CLI)**:
```bash
claude mcp add kubeshark -- kubeshark mcp
```

**Without kubectl access** (direct URL mode):
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

```bash
# Claude Code equivalent:
claude mcp add kubeshark -- kubeshark mcp --url https://kubeshark.example.com
```

## Verification

- Claude Code: `/mcp` to check connection status
- Terminal: `kubeshark mcp --list-tools`
- Cluster: `kubectl get pods -l app=kubeshark-hub`

## Troubleshooting

- **Binary not found** → Install via Homebrew or the install script above
- **Connection refused** → Deploy Kubeshark first: `kubeshark tap`
- **No L7 data** → Check `get_dissection_status` and `enable_dissection`
- **Snapshot creation fails** → Verify raw capture is enabled in Kubeshark config
- **Empty snapshot** → Check `get_data_boundaries` — the requested window may
  fall outside available data
