Do not include any Claude/AI attribution (Co-Authored-By lines, badges, etc.) in commit messages or pull request descriptions.

## Skills

Kubeshark is building an ecosystem of open-source AI skills that work with the Kubeshark MCP.
Skills live in the `skills/` directory at the root of this repo.

### What is a skill?

A skill is a SKILL.md file (with optional reference docs) that teaches an AI agent a domain-specific
methodology. The Kubeshark MCP provides the tools (snapshot creation, API call queries, PCAP export,
etc.) — a skill tells the agent *how* to use those tools for a specific job.

### Skill structure

```
skills/
└── <skill-name>/
    ├── SKILL.md              # Required. YAML frontmatter + markdown instructions.
    └── references/           # Optional. Reference docs loaded on demand.
        └── *.md
```

### SKILL.md format

Every SKILL.md starts with YAML frontmatter:

```yaml
---
name: skill-name
description: >
  When to trigger this skill. Be specific about user intents, keywords, and contexts.
  The description is the primary mechanism for AI agents to decide whether to load the skill.
---
```

The body is markdown instructions that define the methodology: prerequisites, workflows,
tool usage patterns, output guidelines, and reference pointers.

### Guidelines for writing skills

- Keep SKILL.md under 500 lines. Put detailed references in `references/` with clear pointers.
- Use imperative tone ("Check data boundaries", "Create a snapshot").
- Reference Kubeshark MCP tools by exact name (e.g., `create_snapshot`, `list_api_calls`).
- Include realistic example tool responses so the agent knows what to expect.
- Explain *why* things matter, not just *what* to do — the agent is smart and benefits from context.
- Include a Setup Reference section with MCP configuration for Claude Code and Claude Desktop.
- The description frontmatter should be "pushy" — include trigger keywords generously so the skill
  activates when needed. Better to over-trigger than under-trigger.

### Kubeshark MCP tools available to skills

**Cluster management**: `check_kubeshark_status`, `start_kubeshark`, `stop_kubeshark`
**Inventory**: `list_workloads`
**L7 API**: `list_api_calls`, `get_api_call`, `get_api_stats`
**L4 flows**: `list_l4_flows`, `get_l4_flow_summary`
**Snapshots**: `get_data_boundaries`, `create_snapshot`, `get_snapshot`, `list_snapshots`, `start_snapshot_dissection`
**PCAP**: `export_snapshot_pcap`, `resolve_workload`
**Cloud storage**: `get_cloud_storage_status`, `upload_snapshot_to_cloud`, `download_snapshot_from_cloud`
**Dissection**: `get_dissection_status`, `enable_dissection`, `disable_dissection`

### KFL (Kubeshark Filter Language)

KFL2 is built on CEL (Common Expression Language). Skills that involve traffic filtering should
reference KFL. Key concepts:

- Display filter (post-capture), not capture filter
- Fields: `src.ip`, `dst.ip`, `src.pod.name`, `dst.pod.namespace`, `src.service.name`, etc.
- Protocol booleans: `http`, `dns`, `redis`, `kafka`, `tls`, `grpc`, `amqp`, `ws`
- HTTP fields: `url`, `method`, `status_code`, `path`, `request.headers`, `response.headers`,
  `request_body_size`, `response_body_size`, `elapsed_time` (microseconds)
- DNS fields: `dns_questions`, `dns_answers`, `dns_question_types`
- Operators: `==`, `!=`, `<`, `>`, `&&`, `||`, `in`
- String functions: `.contains()`, `.startsWith()`, `.endsWith()`, `.matches()` (regex)
- Collection: `size()`, `[index]`, `[key]`
- Full reference: https://docs.kubeshark.com/en/v2/kfl2

### Key Kubeshark concepts for skill authors

- **eBPF capture**: Kernel-level, no sidecars/proxies. Decrypts TLS without private keys.
- **Protocols**: HTTP, gRPC, GraphQL, WebSocket, Kafka, Redis, AMQP, DNS, and more.
- **Raw capture**: FIFO buffer per node. Must be enabled for retrospective analysis.
- **Snapshots**: Immutable freeze of traffic in a time window. Includes raw capture files,
  K8s pod events, and eBPF cgroup events.
- **Dissection**: The "indexing" step. Reconstructs raw packets into structured L7 API calls.
  Think of it like a search engine indexing web pages — without dissection you have PCAPs,
  with dissection you have a queryable database. Kubeshark is the search engine for network traffic.
- **Cloud storage**: Snapshots can be uploaded to S3/GCS/Azure and downloaded to any cluster.
  A production snapshot can be analyzed on a local KinD cluster.

### Current skills

- `skills/network-rca/` — Network Root Cause Analysis. Retrospective traffic analysis via
  snapshots, dissection, KFL queries, PCAP extraction, trend comparison.
- `skills/kfl/` — KFL2 (Kubeshark Filter Language) expert. Complete reference for writing,
  debugging, and optimizing CEL-based traffic filters across all supported protocols.

### Planned skills (not yet created)

- `skills/api-security/` — OWASP API Top 10 assessment against live or snapshot traffic.
- `skills/incident-response/` — 7-phase forensic incident investigation methodology.
- `skills/network-engineering/` — Real-time traffic analysis, latency debugging, dependency mapping.
