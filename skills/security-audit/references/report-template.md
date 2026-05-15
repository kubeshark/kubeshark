# Security Audit Report Template

Use this template for the markdown report. Fill in all sections, then convert
to PDF.

```markdown
# Kubernetes Network Security Audit Report

**Cluster**: <cluster name/context>
**Namespace**: <target namespace>
**Date**: <audit date and time, local timezone>
**Audit window**: <start time> — <end time> (<duration>)
**Snapshots analyzed**: <count and IDs>
**Audited by**: Claude Code + Kubeshark MCP

---

## Executive Summary

<2-3 sentence summary: how many threats found, highest severity,
whether an active attack chain was identified, top recommendation>

## Threat Summary

| # | Severity | Workload | Threat | MITRE ATT&CK |
|---|----------|----------|--------|---------------|
| 1 | CRITICAL | log-shipper | DNS Tunneling | T1048.003 |
| 2 | CRITICAL | cloud-health-monitor | IMDS Credential Theft | T1552.005 |
| ... | | | | |

## Detailed Findings

### Finding 1: <Title> (CRITICAL)

**Workload**: <pod name>
**MITRE ATT&CK**: <technique ID and name>
**Snapshot**: <snapshot ID>
**Detection method**: <which phase and tool detected this>

**Evidence**:
<Specific traffic data — DNS queries, HTTP requests, L4 flows,
protocol payloads. Include timestamps, source/dest, and relevant
content. Quote actual query names, URLs, SQL statements, or
Redis commands observed.>

**Impact**:
<What this means — data at risk, credentials exposed, scope of access>

**Recommendation**:
<Specific remediation — NetworkPolicy, RBAC change, pod deletion, credential rotation>

---

(repeat for each finding)

## Attack Chain Analysis

<If findings correlate, map the kill chain:
Initial Access → Reconnaissance → Credential Access → Lateral Movement →
Exfiltration → Persistence. Identify which workloads participate in each stage.>

## Detection Coverage

| Phase | Checked | Findings |
|-------|---------|----------|
| Workload Inventory | Yes | <count> |
| DNS Threat Analysis | Yes | <count> |
| External Communication | Yes | <count> |
| Lateral Movement | Yes | <count> |
| Protocol Abuse | Yes | <count> |
| Credential Access | Yes | <count> |

## Limitations

<What this audit cannot detect — config-level vulnerabilities,
image CVEs, idle threats. Recommend complementary tools.>

## Immediate Actions

1. <Highest priority action>
2. <Second priority>
3. ...

## Evidence Preservation

<List snapshot IDs created during this audit. Recommend uploading
to cloud storage for long-term retention. Include PCAP export
commands for key findings.>
```

## Quality Guidelines

- **Include raw evidence** — quote actual DNS queries, HTTP URLs, SQL
  statements, Redis commands. The reader should be able to verify findings
  without re-running the audit.
- **Timestamp everything** — every finding should reference the snapshot ID
  and timestamp (local time with UTC in parentheses).
- **Be specific in recommendations** — not "fix RBAC" but "revoke
  ClusterRoleBinding `mule-recon-cluster-admin` and replace with a
  namespace-scoped Role granting only `get` on `pods`".
- **Include MITRE ATT&CK IDs** — makes the report actionable for security
  teams that track coverage against the framework.
