# Kubeshark Claude Code Plugin

This directory contains the [Claude Code plugin](https://docs.anthropic.com/en/docs/claude-code/plugins) configuration for Kubeshark.

## What's here

| File | Purpose |
|------|---------|
| `plugin.json` | Plugin manifest — name, version, description, metadata |
| `marketplace.json` | Marketplace index — allows discovery via `/plugin marketplace add` |

## Installing the plugin

```
/plugin marketplace add kubeshark/kubeshark
/plugin install kubeshark
```

This loads the Kubeshark AI skills and MCP configuration. Skills appear as
`/kubeshark:network-rca` and `/kubeshark:kfl`.

## What the plugin includes

- **Skills** from [`skills/`](../skills/) — network root cause analysis and KFL filter expertise
- **MCP configuration** from [`.mcp.json`](../.mcp.json) — connects to the Kubeshark MCP server

## Local development

Test the plugin without installing:

```bash
claude --plugin-dir /path/to/kubeshark
```
