<p align="center">
  <img src="https://raw.githubusercontent.com/kubeshark/assets/master/svg/kubeshark-logo.svg" alt="Kubeshark" height="120px"/>
</p>

<p align="center">
    <a href="https://github.com/kubeshark/kubeshark/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/kubeshark/kubeshark?logo=GitHub&style=flat-square"></a>
    <a href="https://hub.docker.com/r/kubeshark/worker"><img alt="Docker pulls" src="https://img.shields.io/docker/pulls/kubeshark/worker?color=%23099cec&logo=Docker&style=flat-square"></a>
    <a href="https://discord.gg/WkvRGMUcx7"><img alt="Discord" src="https://img.shields.io/discord/1042559155224973352?logo=Discord&style=flat-square&label=discord"></a>
    <a href="https://join.slack.com/t/kubeshark/shared_invite/zt-3jdcdgxdv-1qNkhBh9c6CFoE7bSPkpBQ"><img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-green?logo=Slack&style=flat-square"></a>
</p>

<p align="center"><b>Network Observability for SREs & AI Agents</b></p>

<p align="center">
  <a href="https://demo.kubeshark.com/">Live Demo</a> · <a href="https://docs.kubeshark.com">Docs</a>
</p>

---

Kubeshark indexes cluster-wide network traffic at the kernel level using eBPF — delivering instant answers to any query using network, API, and Kubernetes semantics.

**What you can do:**

- **Download Retrospective PCAPs** — cluster-wide packet captures filtered by nodes, time, workloads, and IPs. Store PCAPs for long-term retention and later investigation.
- **Visualize Network Data** — explore traffic matching queries with API, Kubernetes, or network semantics through a real-time dashboard.
- **Integrate with AI** — connect your favorite AI assistant (e.g. Claude, Copilot) to include network data in AI-driven workflows like incident response and root cause analysis.

![Kubeshark](https://github.com/kubeshark/assets/raw/master/png/stream.png)

---

## Get Started

```bash
helm repo add kubeshark https://helm.kubeshark.com
helm install kubeshark kubeshark/kubeshark
kubectl port-forward svc/kubeshark-front 8899:80
```

Open `http://localhost:8899` in your browser. You're capturing traffic.

> For production use, we recommend using an [ingress controller](https://docs.kubeshark.com/en/ingress) instead of port-forward.

**Connect an AI agent** via MCP:

```bash
brew install kubeshark
claude mcp add kubeshark -- kubeshark mcp
```

[MCP setup guide →](https://docs.kubeshark.com/en/mcp)

---

### Network Data for AI Agents

Kubeshark exposes cluster-wide network data via [MCP](https://docs.kubeshark.com/en/mcp) — enabling AI agents to query traffic, investigate API calls, and perform root cause analysis through natural language.

> *"Why did checkout fail at 2:15 PM?"*
> *"Which services have error rates above 1%?"*
> *"Show TCP retransmission rates across all node-to-node paths"*
> *"Trace request abc123 through all services"*

Works with Claude Code, Cursor, and any MCP-compatible AI.

![MCP Demo](https://github.com/kubeshark/assets/raw/master/gif/mcp-demo.gif)

[MCP setup guide →](https://docs.kubeshark.com/en/mcp)

### AI Skills

Open-source, reusable skills that teach AI agents domain-specific workflows on top of Kubeshark's MCP tools:

| Skill | Description |
|-------|-------------|
| **[Network RCA](skills/network-rca/)** | Retrospective root cause analysis — snapshots, dissection, PCAP extraction, trend comparison |
| **[KFL](skills/kfl/)** | KFL (Kubeshark Filter Language) expert — writes, debugs, and optimizes traffic filters |

Install as a Claude Code plugin:

```
/plugin marketplace add kubeshark/kubeshark
/plugin install kubeshark
```

Or clone and use directly — skills trigger automatically based on conversation context.

[AI Skills docs →](https://docs.kubeshark.com/en/mcp/skills)

---

### Query with API, Kubernetes, and Network Semantics

Kubeshark indexes cluster-wide network traffic by parsing it according to protocol specifications, with support for HTTP, gRPC, Redis, Kafka, DNS, and more. A single [KFL query](https://docs.kubeshark.com/en/v2/kfl2) can combine all three semantic layers — Kubernetes identity, API context, and network attributes — to pinpoint exactly the traffic you need. No code instrumentation required.

![KFL query combining API, Kubernetes, and network semantics](https://github.com/kubeshark/assets/raw/master/png/kfl-semantics.png)

[KFL reference →](https://docs.kubeshark.com/en/v2/kfl2) · [Traffic indexing →](https://docs.kubeshark.com/en/v2/l7_api_dissection)

### Workload Dependency Map

A visual map of how workloads communicate, showing dependencies, traffic volume, and protocol usage across the cluster.

![Service Map](https://github.com/kubeshark/assets/raw/master/png/servicemap.png)

[Learn more →](https://docs.kubeshark.com/en/v2/service_map)

### Traffic Retention & PCAP Export

Capture and retain raw network traffic cluster-wide, including decrypted TLS. Download PCAPs scoped by time range, nodes, workloads, and IPs — ready for Wireshark or any PCAP-compatible tool. Store snapshots in cloud storage (S3, Azure Blob, GCS) for long-term retention and cross-cluster sharing.

![Traffic Retention](https://github.com/kubeshark/assets/raw/master/png/snapshots.png)

[Snapshots guide →](https://docs.kubeshark.com/en/v2/traffic_snapshots) · [Cloud storage →](https://docs.kubeshark.com/en/snapshots_cloud_storage)

---

## Features

| Feature | Description |
|---------|-------------|
| [**Traffic Snapshots**](https://docs.kubeshark.com/en/v2/traffic_snapshots) | Point-in-time snapshots with cloud storage (S3, Azure Blob, GCS), PCAP export for Wireshark |
| [**Traffic Indexing**](https://docs.kubeshark.com/en/v2/l7_api_dissection) | Real-time and delayed L7 indexing with request/response matching and full payloads |
| [**Protocol Support**](https://docs.kubeshark.com/en/protocols) | HTTP, gRPC, GraphQL, Redis, Kafka, DNS, and more |
| [**TLS Decryption**](https://docs.kubeshark.com/en/encrypted_traffic) | eBPF-based decryption without key management, included in snapshots |
| [**AI Integration**](https://docs.kubeshark.com/en/mcp) | MCP server + open-source AI skills for network RCA and traffic filtering |
| [**KFL Query Language**](https://docs.kubeshark.com/en/v2/kfl2) | CEL-based query language with Kubernetes, API, and network semantics |
| [**100% On-Premises**](https://docs.kubeshark.com/en/air_gapped) | Air-gapped support, no external dependencies |

---

## Install

| Method | Command |
|--------|---------|
| Helm | `helm repo add kubeshark https://helm.kubeshark.com && helm install kubeshark kubeshark/kubeshark` |
| Homebrew | `brew install kubeshark && kubeshark tap` |
| Binary | [Download](https://github.com/kubeshark/kubeshark/releases/latest) |

[Installation guide →](https://docs.kubeshark.com/en/install)

---

## Contributing

We welcome contributions. See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

[Apache-2.0](LICENSE)
