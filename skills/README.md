# Kubeshark AI Skills

Open-source AI skills that work with the [Kubeshark MCP](https://github.com/kubeshark/kubeshark).
Skills teach AI agents how to use Kubeshark's MCP tools for specific workflows
like root cause analysis, traffic filtering, and forensic investigation.

Skills use the open [Agent Skills](https://github.com/anthropics/skills) format
and work with Claude Code, OpenAI Codex CLI, Gemini CLI, Cursor, and other
compatible agents.

## Available Skills

| Skill | Description |
|-------|-------------|
| [`network-rca`](network-rca/) | Network Root Cause Analysis. Retrospective traffic analysis via snapshots, with two investigation routes: PCAP (for Wireshark/compliance) and Dissection (for AI-driven API-level investigation). |
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

Skills trigger automatically based on your conversation.

### Option 3: Manual installation

Clone the repo (if you haven't already), then symlink or copy the skills:

```bash
git clone https://github.com/kubeshark/kubeshark

# Symlink to stay in sync with the repo (recommended)
ln -s $PWD/kubeshark/skills/network-rca ~/.claude/skills/network-rca
ln -s $PWD/kubeshark/skills/kfl ~/.claude/skills/kfl

# Or copy to your project (project scope only)
mkdir -p .claude/skills
cp -r kubeshark/skills/network-rca .claude/skills/
cp -r kubeshark/skills/kfl .claude/skills/

# Or copy for personal use (all your projects)
cp -r kubeshark/skills/network-rca ~/.claude/skills/
cp -r kubeshark/skills/kfl ~/.claude/skills/
```

### Prerequisites

All skills require the Kubeshark MCP:

```bash
# Claude Code
claude mcp add kubeshark -- kubeshark mcp

# Without kubectl access (direct URL)
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

## Contributing

We welcome contributions — whether improving an existing skill or proposing a new one.

- **Suggest improvements**: Open an issue or PR with changes to an existing skill's `SKILL.md`
  or reference docs. Better examples, clearer workflows, and additional filter patterns
  are always appreciated.
- **Add a new skill**: Open an issue describing the use case first. New skills should
  follow the structure below and reference Kubeshark MCP tools by exact name.

### Skill structure

```
skills/
└── <skill-name>/
    ├── SKILL.md              # Required. YAML frontmatter + markdown body.
    └── references/           # Optional. Detailed reference docs.
        └── *.md
```

### Guidelines

- Keep `SKILL.md` under 500 lines. Use `references/` for detailed content.
- Use imperative tone. Reference MCP tools by exact name.
- Include realistic example tool responses.
- The `description` frontmatter should be generous with trigger keywords.

### Planned skills

- `api-security` — OWASP API Top 10 assessment against live or snapshot traffic.
- `incident-response` — 7-phase forensic incident investigation methodology.
- `network-engineering` — Real-time traffic analysis, latency debugging, dependency mapping.
