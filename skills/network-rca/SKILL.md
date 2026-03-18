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

Dissection reconstructs raw packets into structured L7 API calls — without it,
you have PCAPs; with it, you have a queryable database.

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

**Example `list_api_calls` response** (filtered to `http && status_code >= 500`):
```
┌──────────────────────┬────────┬──────────────────────────┬────────┬───────────┐
│      Timestamp       │ Method │           URL            │ Status │  Elapsed  │
├──────────────────────┼────────┼──────────────────────────┼────────┼───────────┤
│ 2026-03-14 17:23:45  │ POST   │ /api/v1/orders/charge    │ 503    │ 12,340 ms │
│ 2026-03-14 17:23:46  │ POST   │ /api/v1/orders/charge    │ 503    │ 11,890 ms │
│ 2026-03-14 17:23:48  │ GET    │ /api/v1/inventory/check  │ 500    │  8,210 ms │
│ 2026-03-14 17:24:01  │ POST   │ /api/v1/payments/process │ 502    │ 30,000 ms │
└──────────────────────┴────────┴──────────────────────────┴────────┴───────────┘
Src: api-gateway (prod)  →  Dst: payment-service (prod)
```

Use the pattern of repeated failures and high latency to identify the failing
service chain, then drill into individual calls with `get_api_call`.

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

### Other Use Cases

- **Trend analysis** — Take snapshots at regular intervals and compare
  `get_api_stats` across them to detect latency drift, error rate changes,
  or new service-to-service connections.
- **Forensic preservation** — `create_snapshot` + `upload_snapshot_to_cloud`
  for immutable, long-term evidence. Downloadable to any cluster months later.
- **Production-to-local replay** — Upload a production snapshot to cloud,
  download it on a local KinD cluster, and run `start_snapshot_dissection`
  to investigate safely without touching production.

## Setup Reference

For CLI installation, MCP configuration, verification, and troubleshooting,
see `references/setup.md`.
