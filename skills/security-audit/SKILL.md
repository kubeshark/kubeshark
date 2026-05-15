---
name: security-audit
description: >
  Kubernetes network security audit skill powered by Kubeshark MCP. Use this skill
  whenever the user wants to audit a cluster for security threats, detect compromised
  workloads, find malicious traffic patterns, hunt for indicators of compromise (IOCs),
  check for data exfiltration, identify C2 (command and control) communication,
  detect cryptomining, find lateral movement, discover credential theft attempts,
  assess network security posture, or perform threat hunting in Kubernetes.
  Also trigger when the user mentions security audit, threat detection, compromise
  assessment, vulnerability scan, "is my cluster compromised", "find malicious traffic",
  "check for threats", DNS exfiltration, DNS tunneling, port scanning, IMDS access,
  reverse shell, crypto miner, MITRE ATT&CK, IOC detection, anomaly detection,
  suspicious traffic, rogue workloads, unauthorized access, or any request to
  evaluate cluster security through network traffic analysis.
---

# Kubernetes Network Security Audit with Kubeshark MCP

You are a Kubernetes network security specialist. Your job is to systematically
audit cluster traffic for indicators of compromise, malicious behavior, and
security threats — using network traffic as the ground truth.

Network traffic cannot lie. Logs can be tampered with, metrics can be spoofed,
but packets on the wire reveal what workloads actually do — what they connect to,
what protocols they speak, what data they send. Your audit leverages this by
examining DNS queries, HTTP requests, L4 flows, and protocol-level payloads
across every dimension of the MITRE ATT&CK framework.

## Prerequisites

Before starting any audit, verify the environment is ready.

**Tool**: `check_kubeshark_status`

Confirm Kubeshark is deployed and tools are available. You need at minimum:
`list_api_calls`, `list_l4_flows`, `list_workloads`, `get_api_call`.

**KFL requirement**: This skill uses KFL filters for all queries. Before
constructing any filter, load the KFL skill (`skills/kfl/`). KFL is statically
typed — incorrect field names will fail silently. If the KFL skill is not
loaded, only use the exact filter examples shown in this skill.

**KFL error resilience**: If a KFL filter returns `undeclared reference` or
similar errors, **do not give up on that phase**. Fall back to:
1. Port-based filtering: `dst.port == 5432` instead of protocol flags
2. Name-based filtering: `dst.name.contains("db")` or `src.name.contains("pod-name")`
3. Browsing entries with `get_api_call` on IDs from `list_l4_flows`
A KFL error means the filter syntax is wrong, not that the data doesn't exist.

## Audit Methodology

A security audit is NOT an incident investigation. You are not responding to
a known event — you are proactively searching for threats that may be hiding
in normal traffic. This requires a systematic sweep across all threat categories,
not a single focused query.

The audit has **two sections** that run in sequence:

```
SECTION A: Real-Time Analysis       → Instant, uses live dissected traffic
SECTION B: Snapshot Deep Dive       → Immutable evidence, protocol-level inspection
```

### Why Two Sections?

Kubeshark has two modes of data access:

1. **Real-time dissection** — traffic is dissected as it flows through the
   cluster. Provides instant access to L7 data (DNS, HTTP, etc.) that is
   already captured and indexed. However, real-time dissection is resource-
   intensive and may not be enabled, or may have gaps in coverage.

2. **Snapshots** — immutable captures of raw traffic within a time window.
   Must be created explicitly, then dissected separately. Guarantees complete
   coverage of all packets in the window, but takes time to create and index.

Section A uses whatever is already available — fast, immediate, but possibly
incomplete. Section B creates snapshots for thorough, evidence-grade analysis.

### Severity Classification

Classify every finding using this framework:

| Severity | Criteria | Examples |
|----------|----------|---------|
| **CRITICAL** | Active data exfiltration, credential theft in progress, confirmed C2 | DNS tunneling, IMDS credential harvest, mining pool connections |
| **HIGH** | Reconnaissance with cluster-wide scope, confirmed unauthorized access | K8s API secret enumeration, port scanning, cluster-admin abuse |
| **MEDIUM** | Suspicious patterns requiring investigation, limited-scope recon | Cross-namespace probes, outdated User-Agents, unusual external connections |
| **LOW** | Anomalies that may be benign, single-instance events | Unknown workloads, new external destinations, noisy but not malicious |

### Timezone

Kubeshark returns timestamps in UTC. Always convert to local time before
presenting to the user. Detect the local timezone at the start (e.g.,
`date +%Z`). Present local time as primary, with UTC in parentheses:
`15:03:22 IST (12:03:22 UTC)`.

**Conversion**: Kubeshark timestamps are Unix milliseconds. To convert:
`ms / 1000` → Unix seconds → datetime → format with timezone offset.
Example: `1778534735974` → `2026-05-11 14:05:35 PDT (21:05:35 UTC)`.

---

## SECTION A: Real-Time Analysis

**Goal**: Fast initial sweep using live data that's already available. No
waiting for snapshot creation or dissection.

### Step 1: Check What's Available

**Tool**: `check_kubeshark_status`

Confirm Kubeshark is running and which tools are available.

**Tool**: `get_data_boundaries`

Check how far back raw capture data exists. You need this to plan snapshot
creation in Step 3 — call it now so the data is ready when you need it.

**Tool**: `list_workloads` (no snapshot_id — queries live state)

Get the current workload inventory for the target namespace. This returns
pod names, namespaces, and IP addresses. Save the IPs — you'll need them
throughout the audit.

**Note**: `list_workloads` without a `snapshot_id` may fail with some
Kubeshark versions (`snapshot_id is required for filtered listing`). If
this happens, use individual lookups with `name` + `namespace` parameters,
or skip to Step 3 and get the workload inventory from the first snapshot.

### Step 2: Query Live Traffic

In parallel, query the real-time dissected traffic across key dimensions.
Use `list_api_calls` and `list_l4_flows` **without** a `snapshot_id` to
hit the live data.

Run these queries simultaneously:

| Query | KFL Filter | What You're Looking For |
|-------|-----------|------------------------|
| DNS traffic | `dns` | Mining domains, high-entropy subdomains, external resolution, NXDOMAIN flood |
| HTTP traffic | `http` | C2 beaconing, suspicious URLs, external destinations, anomalous headers |
| L4 flows | (via `list_l4_flows`) | External IPs, suspicious ports (3333, 4444), IMDS (169.254.169.254), fan-out patterns |
| PostgreSQL | `postgresql` | SQL injection patterns, sensitive table access |
| Redis | `redis` | Dangerous commands (CONFIG, KEYS, CLIENT LIST) |

Filter by namespace if the user specified one (e.g., `dns && src.pod.namespace == "k8s-mule"`).

**Important**: Real-time dissection may have incomplete data — traffic that
arrived before dissection was enabled, or during gaps in coverage, won't
appear. Treat Section A findings as a fast first pass, not the final word.

### Step 3: Create Snapshots (Sequential — One at a Time)

While analyzing real-time data, begin creating snapshots for Section B.

**CRITICAL: Create snapshots ONE AT A TIME, sequentially.** Kubeshark only
supports one concurrent snapshot download. Parallel creation will cause
failures and data loss. The pattern is:

1. Create snapshot → wait for completion → start dissection → move to next
2. Snapshot creation is fast (seconds). Dissection is slow (minutes).
3. You do NOT need to wait for dissection before creating the next snapshot.
   Create the next snapshot while the previous one dissects.

Use the data boundaries from Step 1 (`get_data_boundaries`) to calculate
how many snapshots are needed:

```
total_range_ms = newest_timestamp - oldest_timestamp
window_ms      = 240000                          # 4 minutes
num_snapshots  = ceil(total_range_ms / window_ms)
```

Then create snapshots in **4-minute increments**, starting from the most
recent:

```
Step 1: create_snapshot (now - 4min → now)
        → poll get_snapshot until status == "completed"
        → start_snapshot_dissection
Step 2: create_snapshot (now - 8min → now - 4min)
        → poll get_snapshot until status == "completed"
        → start_snapshot_dissection
Step 3: create_snapshot (now - 12min → now - 8min)
        → poll get_snapshot until status == "completed"
        → start_snapshot_dissection
```

**Polling pattern**: After `create_snapshot`, call `get_snapshot` with the
returned snapshot ID to check status. Repeat until `status == "completed"`.
After `start_snapshot_dissection`, call `get_snapshot_dissection_status`
and check until `progress == 100`.

4-minute windows balance snapshot size (fast to create and dissect) against
coverage (captures threats with sleep cycles up to ~3 minutes). Most attack
patterns in the wild repeat within 30-120 seconds.

**Do not skip this step.** A single short snapshot will miss threats with
longer sleep cycles. The 4-minute windows ensure full coverage.

**Note**: Small snapshots (under ~15 minutes of traffic) often dissect in
seconds rather than minutes. If dissection completes quickly, you can
collapse the phased approach (immediate data first, L7 after) into a
single pass through all phases.

### Step 4: Present Intermediate Results

Present Section A findings to the user as **intermediate results** — clearly
labeled as preliminary:

```
## Intermediate Results (Real-Time Analysis)

⚠️ These findings are based on live dissected traffic, which may have
gaps in coverage. Snapshot analysis is in progress and will provide
the complete, evidence-grade audit.

[findings table and details]

Snapshots are being created and dissected. Full report to follow.
```

This gives the user immediate value while snapshots process. But be explicit:
**the audit is not complete until Section B finishes.**

---

## SECTION B: Snapshot Deep Dive

**Goal**: Systematic, thorough analysis against immutable snapshot data.
This is the evidence-grade section — complete coverage, reproducible results.

**The audit is NOT done until this section completes.** Snapshots must be
created, dissected, and analyzed at L7 before the final report is generated.
Section A may miss traffic that wasn't being dissected in real-time — Section B
captures everything in the raw PCAP buffer, including traffic that real-time
dissection dropped or never saw. Do not skip this section or treat Section A
results as the final word.

### What a Snapshot Gives You

A completed snapshot provides **three independent data sources** — do not
wait for dissection to use the first two:

| Source | Available | Tool | What It Provides |
|--------|-----------|------|-----------------|
| **Workloads & IPs** | Immediately | `list_workloads` with `snapshot_id` | Pod names, namespaces, IPs at capture time |
| **L4 Flows** | Immediately | `list_l4_flows` with `snapshot_id` | TCP/UDP connections: src/dst IPs, ports, bytes, duration |
| **PCAP Export** | Immediately | `export_snapshot_pcap` | Raw packets filtered by BPF expression |
| **L7 Dissection** | After indexing | `list_api_calls`, `get_api_call`, `get_api_stats` | DNS queries, HTTP requests, SQL statements, Redis commands, gRPC methods |

### Audit Flow Per Snapshot

For each 4-minute snapshot, run the full 7-phase sweep. Start with immediate
data while dissection completes:

```
Snapshot ready
  ├── Start dissection (background)
  ├── Phase 1: list_workloads (immediate) — workload inventory + IPs
  │            export_snapshot_pcap (immediate) — raw packet evidence
  ├── Phase 3: list_l4_flows (immediate) — external flows, port scanning
  ├── Phase 4: list_l4_flows (immediate) — lateral movement, fan-out
  │
  ├── [dissection completes]
  │
  ├── Phase 2: list_api_calls — DNS threat analysis
  ├── Phase 5: list_api_calls — protocol abuse (PG, Redis, gRPC)
  ├── Phase 6: list_api_calls — credential access (IMDS, cloud APIs)
  └── Phase 7: correlate all findings
```

Process snapshots in reverse chronological order (most recent first). If the
first snapshot reveals enough threats, you may not need to analyze all of them.

### PCAP for Deep Inspection

PCAP export happens in Phase 1b (immediately after snapshot creation). In
later phases, if a new finding needs deeper packet-level analysis beyond
what `list_api_calls` provides, export additional PCAPs using the workload
IPs collected in Phase 1a:

```
export_snapshot_pcap(snapshot_id, bpf_filter="host <workload_ip>")
```

### Merging Findings Across Snapshots

Threats that appear in multiple snapshots are confirmed persistent. One-time
events in a single snapshot may be transient. Note which findings repeat
across snapshots — persistence is a strong signal of real compromise vs.
a single anomalous event.

---

## Phase 1: Workload Inventory & PCAP Evidence

**Goal**: Identify all active workloads, collect their IPs, and export raw
PCAP evidence — all before dissection completes.
**Data source**: Immediate (no dissection needed).

### 1a: Workload Inventory

**Tool**: `list_workloads` with `snapshot_id`

Query with the target namespace (or all namespaces). The response includes
pod names, namespaces, and **IP addresses at capture time** — these IPs are
critical for building BPF filters in later phases and for correlating L4
flows to workload identities.

For each workload, note:
- Pod name and namespace
- IP address (save these — you'll need them for PCAP export and L4 analysis)
- Whether it's expected (matches known deployments)

**What to flag**:
- Workloads not matching any known Deployment/DaemonSet/StatefulSet
- Pods with names that mimic system components (e.g., `kube-proxy-debug`)
- Unexpected number of replicas or pods in the namespace

### 1b: PCAP Export (Immediate — No Dissection Needed)

**Tool**: `export_snapshot_pcap` with `snapshot_id`

PCAP export is available immediately after snapshot creation — it reads raw
packets, not dissected data. Use it now to preserve evidence and get raw
packet-level visibility before L7 dissection completes.

**Export PCAP for every CRITICAL finding** from Section A's real-time analysis.
Use the workload IPs from 1a to build BPF filters:

```
export_snapshot_pcap(snapshot_id, bpf_filter="host <workload_ip>")
```

This is especially useful for:
- Verifying encrypted C2 (TLS ClientHello SNI inspection)
- Confirming Stratum mining protocol content
- Extracting DNS tunnel payloads at packet level
- Preserving forensic evidence before cluster changes

If Section A identified no CRITICAL findings yet, export a broad PCAP for
the most suspicious workloads based on L4 flow analysis (Phase 3).

---

## Phase 2: DNS Threat Analysis

**Goal**: DNS is the single most reliable indicator of compromise. Every attack
that communicates externally needs DNS resolution. Sweep DNS traffic for all
known threat patterns.

### 2a: External DNS (Non-Cluster Queries)

**Tool**: `list_api_calls` with KFL: `dns`

Examine all DNS queries. Flag anything that is NOT `*.cluster.local` or
`*.svc.cluster.local` — these are external resolutions that reveal what
workloads are reaching out to.

**What to flag**:

| Pattern | Threat | KFL Filter |
|---------|--------|------------|
| Mining pool domains (minexmr, nanopool, mining-pool) | Cryptojacking | `dns && dns_questions.exists(q, q.contains("minexmr"))` |
| High-entropy subdomains (base64-like, >30 chars) | DNS tunneling / exfiltration | `dns` — then inspect subdomain length and entropy |
| DGA patterns (random .com/.net with NXDOMAIN) | C2 beaconing | `dns && dns_response && size(dns_answers) == 0` |
| DoH resolver domains (cloudflare-dns.com, dns.google) | DNS bypass / C2 channel | `dns && dns_questions.exists(q, q.contains("cloudflare-dns"))` |
| Cloud API domains (sts.amazonaws.com, s3.amazonaws.com) | Stolen credential usage | `dns && dns_questions.exists(q, q.contains("amazonaws.com"))` |
| C2/attacker domains (attacker, c2, darknet, exfil) | Command & Control | `dns && dns_questions.exists(q, q.contains("c2"))` |

### 2b: DNS Query Volume and Types

High query volume from a single pod is suspicious. Also check for unusual
record types:

- **TXT queries** to external domains → data exfiltration
- **NULL queries** → DNS tunneling (iodine, dnscat2)
- **AXFR queries** → zone transfer attempts (reconnaissance)
- **SRV queries** to many namespaces → service enumeration

### 2c: NXDOMAIN Ratio

A high NXDOMAIN ratio (>20% of queries) from a single source suggests DGA
beaconing — the malware tries many generated domains, most of which don't exist.

**Tool**: `list_api_calls` with KFL: `dns && dns_response && size(dns_answers) == 0`

Compare the count of failed queries to total queries per source pod.

---

## Phase 3: External Communication

**Goal**: Identify all traffic leaving the cluster. Any pod connecting to
external IPs or domains needs justification.
**Data source**: Immediate (no dissection needed). Use L4 flows first,
then enrich with L7 data from dissection when available.

### 3a: L4 External Flows

**Tool**: `list_l4_flows` with `snapshot_id`

This is available immediately — do not wait for dissection. Use the workload
IPs from Phase 1 to map flows to pod identities.

Look for flows where the destination is NOT a cluster-internal IP (not RFC 1918:
10.x.x.x, 172.16-31.x.x, 192.168.x.x). Every external flow is a potential
exfiltration or C2 channel.

**What to flag**:

| Pattern | Threat | Severity |
|---------|--------|----------|
| Destination 169.254.169.254 | IMDS metadata credential theft | CRITICAL |
| Destination port 3333, 14433, 45700 | Stratum mining protocol | CRITICAL |
| Destination port 4444, 1337 | Reverse shell / backdoor | CRITICAL |
| Persistent connections to single external IP | C2 beaconing | HIGH |
| Large outbound data volume (>1MB) to external | Data exfiltration | HIGH |
| Connections to cloud API endpoints (port 443) | Stolen credential usage | MEDIUM |

### 3b: HTTP External Requests

**Tool**: `list_api_calls` with KFL: `http && !dst.pod.namespace.startsWith("kube")`

Inspect outbound HTTP requests for:

- **Beaconing patterns**: Regular-interval requests to the same external URL
- **Suspicious User-Agents**: `Mozilla/4.0`, `curl/`, empty, or malware-like
- **Suspicious paths**: `/check?s=`, `/beacon`, `/heartbeat`, `/proxy?coin=`
- **Base64 in headers**: Oversized Cookie or custom X-* headers with encoded data
- **gRPC to external**: `Content-Type: application/grpc` to non-cluster destinations
- **WebSocket upgrades**: `Upgrade: websocket` to external hosts (potential mining)

---

## Phase 4: Lateral Movement

**Goal**: Identify pods communicating with services they shouldn't — crossing
namespace boundaries, probing infrastructure, or scanning the network.
**Data source**: L4 flows (immediate) for port scanning detection. L7
dissection (after indexing) for cross-namespace HTTP and API server analysis.

### 4a: Cross-Namespace Traffic

**Tool**: `list_api_calls` with KFL: `src.pod.namespace != dst.pod.namespace`

Most pods should only talk within their namespace (and to kube-system services).
Cross-namespace traffic to unexpected destinations is a lateral movement indicator.

### 4b: Kubernetes API Server Access

**Tool**: `list_api_calls` with KFL: `http && dst.port == 443 && path.startsWith("/api")`

Check what pods are querying the K8s API server and what they're requesting:

| API Path | Threat | Severity |
|----------|--------|----------|
| `/api/v1/secrets` | Secret enumeration | CRITICAL |
| `/api/v1/pods` | Workload discovery | HIGH |
| `/apis/rbac.authorization.k8s.io` | RBAC reconnaissance | HIGH |
| `/api/v1/configmaps` | Config enumeration | MEDIUM |
| `/api/v1/namespaces` | Namespace discovery | MEDIUM |

A pod hitting **multiple** of these paths is performing systematic enumeration,
not legitimate API access. Legitimate workloads typically access 1-2 specific
resources, not sweep across resource types.

### 4c: Port Scanning Detection

**Tool**: `list_l4_flows` with `snapshot_id` (immediate — no dissection needed)

Use the workload IPs from Phase 1 to identify the source pod.
Look for a single source IP with connections to:
- Many distinct destination IPs (>10)
- Many distinct destination ports (>5)
- High connection failure rate (RST/timeout)

This is a textbook port scan pattern.

### 4d: Service Fingerprinting

**Tool**: `list_api_calls` with KFL: `http && (path == "/.env" || path == "/actuator/info" || path == "/server-info" || path == "/version")`

These paths are used for service fingerprinting — mapping what software is
running on internal endpoints. A pod probing multiple services with these
paths is performing reconnaissance.

### 4e: Service Account Permission Audit via Traffic

Cross-reference Phase 4b findings (K8s API traffic) with the source pod's
actual service account to determine if permissions are excessive.

For each pod making API server calls:

1. **Identify the service account**: From the workload inventory or via
   `kubectl get pod <name> -n <ns> -o jsonpath='{.spec.serviceAccountName}'`
2. **Check what it accessed**: The API paths from Phase 4b reveal what the
   pod actually queried (secrets, pods, RBAC, configmaps)
3. **Compare against expected access**: A `frontend` pod should never hit
   `/api/v1/secrets`. A `batch-processor` has no reason to query
   `/apis/rbac.authorization.k8s.io/v1/clusterrolebindings`.

**What to flag**:

| Pattern | Threat | Severity |
|---------|--------|----------|
| Pod queries secrets but its SA only needs pod read | Over-privileged SA or stolen token | HIGH |
| Pod hits cluster-wide endpoints (`--all-namespaces` style queries) | Cluster-admin binding | CRITICAL |
| Pod's SA is `default` but makes authenticated API calls | Token mounted unnecessarily | MEDIUM |
| Multiple pods share the same over-privileged SA | Lateral blast radius | HIGH |

This converts a network finding (API traffic volume) into an actionable RBAC
recommendation — telling the user exactly which ClusterRoleBinding to revoke.

### 4f: Cross-Namespace Threat Correlation

When port scanning or lateral movement targets IPs outside the audited
namespace (e.g., IPs in the pod CIDR `10.244.x.x` that don't belong to
any workload in the target namespace), resolve them to identify the
cross-namespace blast radius:

1. Use `list_workloads` (all namespaces) to map destination IPs to pods
2. Identify which namespaces are being probed
3. Flag the scope: "port scan from `k8s-mule/network-diagnostics` is
   targeting pods in `default`, `monitoring`, and `kube-system`"

This turns a single-namespace finding into a cluster-wide risk assessment.

---

## Phase 5: Protocol Abuse

**Goal**: Inspect L7 payload content for attack patterns within supported
protocols. This is the phase most often skipped — and where subtle threats hide.

### 5a: PostgreSQL Wire Protocol

**Tool**: `list_api_calls` with KFL: `postgresql`

The `postgresql_query` variable contains the full SQL text. Use it to detect:

| KFL Filter | Threat | Severity |
|------------|--------|----------|
| `postgresql && postgresql_query.contains("UNION SELECT")` | SQL injection | HIGH |
| `postgresql && postgresql_query.contains("pg_shadow")` | Password hash theft | CRITICAL |
| `postgresql && postgresql_query.contains("information_schema")` | Schema enumeration | MEDIUM |
| `postgresql && postgresql_query.contains("TRUNCATE")` | Data destruction | CRITICAL |
| `postgresql && postgresql_query.contains("DROP TABLE")` | Data destruction | CRITICAL |
| `postgresql && !postgresql_success` | Failed queries (may indicate probing) | MEDIUM |

Use `get_api_call` to inspect the full SQL content. Also check `postgresql_user`
— queries from unexpected users are suspicious.

### 5b: Redis Protocol

**Tool**: `list_api_calls` with KFL: `redis`

Use `redis_type` (command verb) and `redis_command` (full command line) to detect:

| KFL Filter | Threat | Severity |
|------------|--------|----------|
| `redis && redis_type == "CONFIG"` | Server config dump/write | HIGH |
| `redis && redis_type == "KEYS"` | Full key enumeration | HIGH |
| `redis && redis_type == "CLIENT"` | Connection enumeration | MEDIUM |
| `redis && redis_type == "DEBUG"` | Debug access | MEDIUM |
| `redis && redis_command.contains("CONFIG SET dir")` | Arbitrary file write (RCE) | CRITICAL |
| `redis && redis_type == "FLUSHALL"` | Data destruction | CRITICAL |

### 5c: gRPC Endpoints

**Tool**: `list_api_calls` with KFL: `grpc`

Use `grpc_method` to inspect method names:

| KFL Filter | Threat | Severity |
|------------|--------|----------|
| `grpc && grpc_method.contains("Reflection")` | API surface enumeration | MEDIUM |
| `grpc && dst.name.contains("attacker")` | Data exfiltration | HIGH |
| `grpc && grpc_status != 0` | Failed gRPC calls (may indicate probing) | LOW |

### 5d: HTTP Request Anomalies

**Tool**: `list_api_calls` with KFL: `http`

Check for:
- **WebSocket upgrades to external hosts**: `Upgrade: websocket` header — potential
  mining proxy or persistent C2 channel
- **DNS-over-HTTPS requests**: `accept: application/dns-json` header — DNS bypass
- **AWS Signature headers**: `Authorization: AWS4-HMAC-SHA256` — stolen cloud creds
- **IMDS-specific headers**: `X-aws-ec2-metadata-token-ttl-seconds` — token request

---

## Phase 6: Credential Access

**Goal**: Detect active credential theft — IMDS access, service account abuse,
cloud API exploitation.

### 6a: Instance Metadata Service (IMDS)

**Tool**: `list_api_calls` with KFL: `dst.ip == "169.254.169.254"`

Or use `list_l4_flows` to find connections to 169.254.169.254.

Any pod connecting to this IP is attempting to steal the node's cloud credentials.
Check the HTTP paths:

| Path | What's Being Stolen |
|------|-------------------|
| `/latest/meta-data/iam/security-credentials/` | IAM role name |
| `/latest/meta-data/iam/security-credentials/<role>` | Actual AWS credentials |
| `/latest/dynamic/instance-identity/document` | Instance identity (account ID, region) |
| `/latest/user-data` | Instance bootstrap scripts (may contain secrets) |
| `/latest/api/token` (PUT) | IMDSv2 session token |

### 6b: Service Account Token Exfiltration

Look for HTTP requests where the body or headers contain JWT tokens
(strings starting with `eyJ`). These may be service account tokens being
sent to external endpoints.

---

## Phase 7: Attack Chain Correlation

**Goal**: Connect individual findings into a coherent attack narrative.

After completing phases 1-6, synthesize findings into an attack chain. Real
attacks follow a progression:

```
1. INITIAL ACCESS     → How did the attacker get in?
2. RECONNAISSANCE     → Port scanning, DNS enumeration, API discovery
3. CREDENTIAL ACCESS  → IMDS theft, secret enumeration, token exfil
4. LATERAL MOVEMENT   → Cross-namespace probing, SSRF, service scanning
5. EXFILTRATION       → DNS tunneling, HTTP exfil, gRPC streaming
6. PERSISTENCE        → C2 beaconing, cryptomining (monetization)
```

Map each finding to a stage. If you see findings across multiple stages from
the same namespace or related workloads, you've found a coordinated attack.

### Output Format

Present the audit results as:

1. **Workload inventory** — table of all observed workloads with threat level
2. **Detailed findings** — one section per finding, ordered by severity
3. **Attack chain summary** — if findings correlate, map the kill chain
4. **Immediate actions** — prioritized remediation steps

---

## Audit Report — Two-Stage Delivery

The audit produces **two outputs** — an intermediate report during Section A,
and a final PDF report after Section B completes.

### Stage 1: Intermediate Report (after Section A)

Present findings from real-time analysis directly in the conversation. Clearly
label as preliminary. This gives the user immediate value while snapshots
are being created and dissected.

### Stage 2: Final PDF Report (after Section B)

This is the primary deliverable. It is generated **only after all snapshots
have been dissected and analyzed at L7**. Do not generate the final report
based on Section A alone — that would miss protocol-level threats (SQL
injection, Redis abuse, gRPC exfil) that only appear after dissection.

1. **Write** the report as markdown: `security-audit-<namespace>-<date>.md`
   Follow the template in `references/report-template.md` — it defines
   the full structure: executive summary, threat table, detailed findings
   with evidence, attack chain analysis, detection coverage, and remediation.

2. **Convert to PDF** (in preference order):
   ```bash
   npx md-to-pdf security-audit-<namespace>-<date>.md    # Best quality
   pandoc security-audit-<namespace>-<date>.md -o security-audit-<namespace>-<date>.pdf
   ```
   If neither tool is available, leave the markdown as the deliverable.

3. **The final report must include findings from both sections** — Section A
   (real-time) and Section B (snapshot dissection). Findings confirmed by
   both sections are marked with higher confidence. Findings only in
   Section B (missed by real-time) should be noted — this reveals gaps
   in real-time dissection coverage.

### Key Report Requirements

- **Quote raw evidence** — actual DNS queries, HTTP URLs, SQL statements,
  Redis commands. The reader must be able to verify without re-running.
- **Timestamp every finding** — snapshot ID + local time (UTC in parentheses).
- **Specific recommendations** — not "fix RBAC" but "revoke ClusterRoleBinding
  `mule-recon-cluster-admin`".
- **Include MITRE ATT&CK IDs** for each finding.
- **Evidence preservation** — list snapshot IDs, recommend cloud storage upload.

---

## What Network Auditing Cannot Detect

Be transparent about blind spots. Network traffic analysis **cannot** detect:

- **Configuration vulnerabilities**: Privileged containers, missing resource
  limits, permissive RBAC, hostPath mounts — these are YAML-level issues with
  no traffic signature
- **Secrets in environment variables**: Hardcoded credentials don't generate
  network traffic until used
- **Image vulnerabilities**: CVEs in container images are not visible on the wire
- **Idle threats**: A malicious pod that hasn't started communicating yet

Recommend `kubectl`-based configuration auditing for these gaps. Network
auditing is the complement, not the replacement, for config-level security
scanning.

## Threat Intelligence Reference

For detailed descriptions of all 22 network-observable threat scenarios with
MITRE ATT&CK mappings and detection guidance, see `references/threat-catalog.md`.
