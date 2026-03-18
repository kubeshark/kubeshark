# Kubeshark AI Skills

Open-source AI skills that work with the [Kubeshark MCP](https://github.com/kubeshark/kubeshark).

## What is a skill?

A skill is a `SKILL.md` file (with optional reference docs) that teaches an AI agent a domain-specific methodology. The Kubeshark MCP provides the tools (snapshot creation, API call queries, PCAP export, etc.) — a skill tells the agent *how* to use those tools for a specific job.

## Available Skills

| Skill | Description |
|-------|-------------|
| [`network-rca`](network-rca/) | Network Root Cause Analysis. Retrospective traffic analysis via snapshots, dissection, KFL queries, PCAP extraction, and trend comparison. |
| [`kfl`](kfl/) | KFL2 (Kubeshark Filter Language) expert. Complete reference for writing, debugging, and optimizing CEL-based traffic filters across all supported protocols. |

## Installation

### Option 1: Plugin (recommended)

Install as a Claude Code plugin directly from GitHub:

```
/plugin marketplace add kubeshark/kubeshark
/plugin install kubeshark
```

Skills appear as `/kubeshark:network-rca` and `/kubeshark:kfl`. The plugin
also bundles the Kubeshark MCP configuration automatically.

### Option 2: Clone and run

```bash
git clone https://github.com/kubeshark/kubeshark
cd kubeshark
claude
```

Skills trigger automatically based on your conversation — ask about root cause
analysis, traffic filtering, snapshots, or KFL and the relevant skill loads.

### Option 3: Manual installation

Copy or symlink individual skills into your Claude Code skills directory:

```bash
# Project scope (this project only)
mkdir -p .claude/skills
cp -r skills/network-rca .claude/skills/
cp -r skills/kfl .claude/skills/

# Personal scope (all your projects)
cp -r skills/network-rca ~/.claude/skills/
cp -r skills/kfl ~/.claude/skills/

# Or symlink to stay in sync with the repo
ln -s /path/to/kubeshark/skills/network-rca ~/.claude/skills/network-rca
ln -s /path/to/kubeshark/skills/kfl ~/.claude/skills/kfl
```

Skills are discovered automatically — no restart needed.

### Compatibility

Skills use the open [Agent Skills](https://github.com/anthropics/skills) format
(`SKILL.md` with YAML frontmatter) and work with any compatible agent, including
Claude Code, OpenAI Codex CLI, Gemini CLI, and Cursor.

### Prerequisites

All skills require the Kubeshark MCP to be configured:

```bash
# Claude Code
claude mcp add kubeshark -- kubeshark mcp

# Or with a direct URL (no kubectl required)
claude mcp add kubeshark -- kubeshark mcp --url https://kubeshark.example.com
```

For Claude Desktop, add to `claude_desktop_config.json`:

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

## Skill structure

```
skills/
└── <skill-name>/
    ├── SKILL.md              # Required. YAML frontmatter + markdown instructions.
    └── references/           # Optional. Reference docs loaded on demand.
        └── *.md
```

Every `SKILL.md` starts with YAML frontmatter:

```yaml
---
name: skill-name
description: >
  When to trigger this skill. Be specific about user intents, keywords, and contexts.
  The description is the primary mechanism for AI agents to decide whether to load the skill.
---
```

The body is markdown instructions that define the methodology: prerequisites, workflows, tool usage patterns, and reference pointers.

## Contributing a skill

We welcome contributions! Here are the guidelines:

- Keep `SKILL.md` under 500 lines. Put detailed references in `references/` with clear pointers.
- Use imperative tone ("Check data boundaries", "Create a snapshot").
- Reference Kubeshark MCP tools by exact name (e.g., `create_snapshot`, `list_api_calls`).
- Include realistic example tool responses so the agent knows what to expect.
- Explain *why* things matter, not just *what* to do — the agent benefits from context.
- Include a Setup Reference section with MCP configuration for Claude Code and Claude Desktop.
- The `description` frontmatter should be generous with trigger keywords — better to over-trigger than under-trigger.

### Available MCP tools

Skills can use these Kubeshark MCP tools:

| Category | Tools |
|----------|-------|
| Cluster management | `check_kubeshark_status`, `start_kubeshark`, `stop_kubeshark` |
| Inventory | `list_workloads` |
| L7 API | `list_api_calls`, `get_api_call`, `get_api_stats` |
| L4 flows | `list_l4_flows`, `get_l4_flow_summary` |
| Snapshots | `get_data_boundaries`, `create_snapshot`, `get_snapshot`, `list_snapshots`, `start_snapshot_dissection` |
| PCAP | `export_snapshot_pcap`, `resolve_workload` |
| Cloud storage | `get_cloud_storage_status`, `upload_snapshot_to_cloud`, `download_snapshot_from_cloud` |
| Dissection | `get_dissection_status`, `enable_dissection`, `disable_dissection` |

## Planned skills

- `api-security` — OWASP API Top 10 assessment against live or snapshot traffic.
- `incident-response` — 7-phase forensic incident investigation methodology.
- `network-engineering` — Real-time traffic analysis, latency debugging, dependency mapping.
