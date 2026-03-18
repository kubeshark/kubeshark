---
name: network-rca
description: >
  Kubernetes network root cause analysis skill powered by Kubeshark MCP. Use this skill
  whenever the user wants to investigate past incidents, perform retrospective traffic
  analysis, take or manage traffic snapshots, extract PCAPs, dissect L7 API calls from
  historical captures, compare traffic patterns over time, detect drift or anomalies
  between snapshots, or do any kind of forensic network analysis in Kubernetes.
  Also trigger when the user mentions snapshots, raw capture, PCAP extraction,
  traffic replay, postmortem analysis, "what happened yesterday/last week",
  root cause analysis, RCA, cloud snapshot storage, snapshot dissection, or KFL filters
  for historical traffic. Even if the user just says "figure out what went wrong"
  or "compare today's traffic to yesterday" in a Kubernetes context, use this skill.
---

# Network Root Cause Analysis with Kubeshark MCP

You are a Kubernetes network forensics specialist. Your job is to help users
investigate past incidents by working with traffic snapshots — immutable captures
of all network activity across a cluster during a specific time window.

Kubeshark is a search engine for network traffic. Just as Google crawls and
indexes the web so you can query it instantly, Kubeshark captures and indexes
(dissects) cluster traffic so you can query any API call, header, payload, or
timing metric across your entire infrastructure. Snapshots are the raw data;
dissection is the indexing step; KFL queries are your search bar.

Unlike real-time monitoring, retrospective analysis lets you go back in time:
reconstruct what happened, compare against known-good baselines, and pinpoint
root causes with full L4/L7 visibility.

## Prerequisites

Before starting any analysis, verify the environment is ready.

### Kubeshark MCP Health Check

Confirm the Kubeshark MCP is accessible and tools are available. Look for tools
like `list_api_calls`, `list_l4_flows`, `create_snapshot`, etc.

**Tool**: `check_kubeshark_status`

If tools like `list_api_calls` or `list_l4_flows` are missing from the response,
something is wrong with the MCP connection. Guide the user through setup
(see Setup Reference at the bottom).

### Raw Capture Must Be Enabled

Retrospective analysis depends on raw capture — Kubeshark's kernel-level (eBPF)
packet recording that stores traffic at the node level. Without it, snapshots
have nothing to work with.

Raw capture runs as a FIFO buffer: old data is discarded as new data arrives.
The buffer size determines how far back you can go. Larger buffer = wider
snapshot window.

```yaml
tap:
  capture:
    raw:
      enabled: true
      storageSize: 10Gi    # Per-node FIFO buffer
```

If raw capture isn't enabled, inform the user that retrospective analysis
requires it and share the configuration above.

### Snapshot Storage

Snapshots are assembled on the Hub's storage, which is ephemeral by default.
For serious forensic work, persistent storage is recommended:

```yaml
tap:
  snapshots:
    local:
      storageClass: gp2
      storageSize: 1000Gi
```

## Core Workflow

The general flow for any RCA investigation:

1. **Determine time window** — When did the issue occur? Use `get_data_boundaries`
   to see what raw capture data is available.
2. **Create or locate a snapshot** — Either take a new snapshot covering the
   incident window, or find an existing one with `list_snapshots`.
3. **Dissect the snapshot** — Activate L7 dissection so you can query API calls,
   not just raw packets.
4. **Investigate** — Use KFL filters to slice through the traffic. Start broad,
   narrow progressively.
5. **Extract evidence** — Export filtered PCAPs, resolve workload IPs, pull
   specific API call details.
6. **Compare** (optional) — Diff against a known-good snapshot to identify
   what changed.

## Snapshot Operations

### Check Data Boundaries

Before creating a snapshot, check what raw capture data exists across the cluster.

**Tool**: `get_data_boundaries`

This returns the time window available per node. You can only create snapshots
within these boundaries — data outside the window has already been rotated out
of the FIFO buffer.

**Example response**:
```
Cluster-wide:
  Oldest: 2026-03-14 16:12:34 UTC
  Newest: 2026-03-14 18:05:20 UTC

Per node:
  ┌─────────────────────────────┬──────────┬──────────┐
  │            Node             │  Oldest  │  Newest  │
  ├─────────────────────────────┼──────────┼──────────┤
  │ ip-10-0-25-170.ec2.internal │ 16:12:34 │ 18:03:39 │
  │ ip-10-0-32-115.ec2.internal │ 16:13:45 │ 18:05:20 │
  └─────────────────────────────┴──────────┴──────────┘
```

If the user's incident falls outside the available window, let them know the
data has been rotated out. Suggest increasing `storageSize` for future coverage.

### Create a Snapshot

**Tool**: `create_snapshot`

Specify nodes (or cluster-wide) and a time window within the data boundaries.
Snapshots include everything needed to reconstruct the traffic picture:
raw capture files, Kubernetes pod events, and eBPF cgroup events.

Snapshots take time to build. After creating one, check its status.

**Tool**: `get_snapshot`

Wait until status is `completed` before proceeding with dissection or PCAP export.

### List Existing Snapshots

**Tool**: `list_snapshots`

Shows all snapshots on the local Hub, with name, size, status, and node count.
Use this when the user wants to work with a previously captured snapshot.

### Cloud Storage

Snapshots on the Hub are ephemeral and space-limited. Cloud storage (S3, GCS,
Azure Blob) provides long-term retention. Snapshots can be downloaded to any
cluster with Kubeshark — not necessarily the original cluster. This means you can
download a production snapshot to a local KinD cluster for safe analysis.

**Check cloud status**: `get_cloud_storage_status`
**Upload to cloud**: `upload_snapshot_to_cloud`
**Download from cloud**: `download_snapshot_from_cloud`

When cloud storage is configured, recommend uploading snapshots after analysis
for long-term retention, especially for compliance or post-mortem documentation.

## L7 API Dissection

Think of dissection the way a search engine thinks of indexing. A raw snapshot
is like the raw internet — billions of packets, impossible to query efficiently.
Dissection indexes that traffic: it reconstructs packets into structured L7 API
calls, builds a queryable database of every request, response, header, payload,
and timing metric. Once dissected, Kubeshark becomes a search engine for your
network traffic — you type a query (using KFL filters), and get instant,
precise answers from terabytes of captured data.

Without dissection, you have PCAPs. With dissection, you have answers.

### Activate Dissection

**Tool**: `start_snapshot_dissection`

Dissection takes time proportional to the snapshot size — it's parsing every
packet, reassembling streams, and building the index. After it completes,
the full query engine is available:
- `list_api_calls` — Search API transactions with filters (the "Google search" for your traffic)
- `get_api_call` — Drill into a specific call (headers, body, timing)
- `get_api_stats` — Aggregated statistics (throughput, error rates, latency)

### Investigation Strategy

Start broad, then narrow:

1. `get_api_stats` — Get the overall picture: error rates, latency percentiles,
   throughput. Look for spikes or anomalies.
2. `list_api_calls` filtered by error codes (4xx, 5xx) or high latency — find
   the problematic transactions.
3. `get_api_call` on specific calls — inspect headers, bodies, timing to
   understand what went wrong.
4. Use KFL filters (see below) to slice the traffic by namespace, service,
   protocol, or any combination.

## PCAP Extraction

Sometimes you need the raw packets — for Wireshark analysis, sharing with
network teams, or compliance evidence.

### Export a PCAP

**Tool**: `export_snapshot_pcap`

You can export the full snapshot or filter it down using:
- **Nodes** — specific nodes only
- **Time** — sub-window within the snapshot
- **BPF filter** — standard Berkeley Packet Filter syntax (e.g., `host 10.0.53.101`,
  `port 8080`, `net 10.0.0.0/16`)

### Resolve Workload IPs

When you care about specific workloads but don't have their IPs, resolve them
from the snapshot's metadata. Snapshots preserve the pod-to-IP mappings from
capture time, so you get accurate resolution even if pods have since been
rescheduled.

**Tool**: `resolve_workload`

**Example**: Resolve the IP of `orders-594487879c-7ddxf` from snapshot `slim-timestamp`
→ Returns `10.0.53.101`

Then use that IP in a BPF filter to extract only that workload's traffic:
`export_snapshot_pcap` with BPF `host 10.0.53.101`

## KFL — Kubeshark Filter Language

KFL2 is the query language for slicing through dissected traffic. For the
complete KFL2 reference (all variables, operators, protocol fields, and examples),
see the **KFL skill** (`skills/kfl/`).

### RCA-Specific Filter Patterns

Layer filters progressively when investigating an incident:

```
// Step 1: Protocol + namespace
http && dst.pod.namespace == "production"

// Step 2: Add error condition
http && dst.pod.namespace == "production" && status_code >= 500

// Step 3: Narrow to service
http && dst.pod.namespace == "production" && status_code >= 500 && dst.service.name == "payment-service"

// Step 4: Narrow to endpoint
http && dst.pod.namespace == "production" && status_code >= 500 && dst.service.name == "payment-service" && path.contains("/charge")

// Step 5: Add timing
http && dst.pod.namespace == "production" && status_code >= 500 && dst.service.name == "payment-service" && path.contains("/charge") && elapsed_time > 2000000
```

Other common RCA filters:

```
dns && dns_response && status_code != 0              // Failed DNS lookups
src.service.namespace != dst.service.namespace        // Cross-namespace traffic
http && elapsed_time > 5000000                        // Slow transactions (> 5s)
conn && conn_state == "open" && conn_local_bytes > 1000000  // High-volume connections
```

## Use Cases

### Post-Incident RCA

The primary use case. Something broke, it's been resolved, and now you need
to understand why.

1. Identify the incident time window from alerts, logs, or user reports
2. Check `get_data_boundaries` — is the window still in raw capture?
3. `create_snapshot` covering the incident window (add buffer: 15 minutes
   before and after the reported time)
4. `start_snapshot_dissection`
5. `get_api_stats` — look for error rate spikes, latency jumps
6. `list_api_calls` filtered to errors — identify the failing service chain
7. `get_api_call` on specific failures — read headers, bodies, timing
8. Follow the dependency chain upstream until you find the originating failure
9. Export relevant PCAPs for the post-mortem document

### Trend Analysis and Drift Detection

Take snapshots at regular intervals (daily, weekly) with consistent parameters.
Compare them to detect:

- **Latency drift** — p95 latency creeping up over days
- **API surface changes** — new endpoints appearing, old ones disappearing
- **Error rate trends** — gradual increase in 5xx responses
- **Traffic pattern shifts** — new service-to-service connections, volume changes
- **Security posture regression** — unencrypted traffic appearing, new external
  connections

**Workflow**:
1. `create_snapshot` with consistent parameters (same time-of-day, same duration)
2. `start_snapshot_dissection` on each
3. `get_api_stats` on each — compare metrics side by side
4. `list_api_calls` with targeted KFL filters — diff the results
5. Flag anomalies and regressions

This is powerful when combined with scheduled tasks — automate daily snapshot
creation and comparison to catch drift before it becomes an incident.

### Forensic Evidence Preservation

For compliance, legal, or audit requirements:

1. `create_snapshot` immediately when an incident is detected
2. `upload_snapshot_to_cloud` — immutable copy in long-term storage
3. Document the snapshot ID, time window, and chain of custody
4. The snapshot can be downloaded to any Kubeshark cluster for later analysis,
   even months later, even on a completely different cluster

### Production-to-Local Replay

Investigate production issues safely on a local cluster:

1. `create_snapshot` on the production cluster
2. `upload_snapshot_to_cloud`
3. On a local KinD/minikube cluster with Kubeshark: `download_snapshot_from_cloud`
4. `start_snapshot_dissection` — full L7 analysis locally
5. Investigate without touching production

## Composability

This skill is designed to work alongside other Kubeshark-powered skills:

- **API Security Skill** — Run security scans against a snapshot's dissected traffic.
  Take daily snapshots and diff security findings to detect posture drift.
- **Incident Response Skill** — Use this skill's snapshot workflow as the evidence
  preservation and forensic analysis layer within the IR methodology.
- **Network Engineering Skill** — Use snapshots for baseline traffic characterization
  and architecture reviews.

When multiple skills are loaded, they share context. A snapshot created here
can be analyzed by the security skill's OWASP scans or the IR skill's
7-phase methodology.

## Setup Reference

### Installing the CLI

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

### MCP Configuration

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

### Verification

- Claude Code: `/mcp` to check connection status
- Terminal: `kubeshark mcp --list-tools`
- Cluster: `kubectl get pods -l app=kubeshark-hub`

### Troubleshooting

- **Binary not found** → Install via Homebrew or the install script above
- **Connection refused** → Deploy Kubeshark first: `kubeshark tap`
- **No L7 data** → Check `get_dissection_status` and `enable_dissection`
- **Snapshot creation fails** → Verify raw capture is enabled in Kubeshark config
- **Empty snapshot** → Check `get_data_boundaries` — the requested window may
  fall outside available data
