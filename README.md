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

Kubeshark captures cluster-wide network traffic at the speed and scale of Kubernetes, continuously, at the kernel level using eBPF. It consolidates a highly fragmented picture — dozens of nodes, thousands of workloads, millions of connections — into a single, queryable view with full Kubernetes and API context.

Network data is available to **AI agents via [MCP](https://docs.kubeshark.com/en/mcp)** and to **human operators via a [dashboard](https://docs.kubeshark.com/en/v2)**.

**What's captured, cluster-wide:**

- **L4 Packets & TCP Metrics** — retransmissions, RTT, window saturation, connection lifecycle, packet loss across every node-to-node path ([TCP insights →](https://docs.kubeshark.com/en/mcp/tcp_insights))
- **L7 API Calls** — real-time request/response matching with full payload parsing: HTTP, gRPC, GraphQL, Redis, Kafka, DNS ([API dissection →](https://docs.kubeshark.com/en/v2/l7_api_dissection))
- **Decrypted TLS** — eBPF-based TLS decryption without key management
- **Kubernetes Context** — every packet and API call resolved to pod, service, namespace, and node
- **PCAP Retention** — point-in-time raw packet snapshots, exportable for Wireshark ([Snapshots →](https://docs.kubeshark.com/en/v2/traffic_snapshots))

![Kubeshark](https://github.com/kubeshark/assets/raw/master/png/stream.png)

---

## Get Started

```bash
helm repo add kubeshark https://helm.kubeshark.com
helm install kubeshark kubeshark/kubeshark
```

Dashboard opens automatically. You're capturing traffic.

**Connect an AI agent** via MCP:

```bash
brew install kubeshark
claude mcp add kubeshark -- kubeshark mcp
```

[MCP setup guide →](https://docs.kubeshark.com/en/mcp)

---

### AI-Powered Network Analysis

Kubeshark exposes all cluster-wide network data via MCP (Model Context Protocol). AI agents can query L4 metrics, investigate L7 API calls, analyze traffic patterns, and run root cause analysis — through natural language. Use cases include incident response, root cause analysis, troubleshooting, debugging, and reliability workflows.

> *"Why did checkout fail at 2:15 PM?"*
> *"Which services have error rates above 1%?"*
> *"Show TCP retransmission rates across all node-to-node paths"*
> *"Trace request abc123 through all services"*

Works with Claude Code, Cursor, and any MCP-compatible AI.

![MCP Demo](https://github.com/kubeshark/assets/raw/master/gif/mcp-demo.gif)

[MCP setup guide →](https://docs.kubeshark.com/en/mcp)

---

### L7 API Dissection

Cluster-wide request/response matching with full payloads, parsed according to protocol specifications. HTTP, gRPC, Redis, Kafka, DNS, and more. Every API call resolved to source and destination pod, service, namespace, and node. No code instrumentation required.

![API context](https://github.com/kubeshark/assets/raw/master/png/api_context.png)

[Learn more →](https://docs.kubeshark.com/en/v2/l7_api_dissection)

### L4/L7 Workload Map

Cluster-wide view of service communication: dependencies, traffic flow, and anomalies across all nodes and namespaces.

![Service Map](https://github.com/kubeshark/assets/raw/master/png/servicemap.png)

[Learn more →](https://docs.kubeshark.com/en/v2/service_map)

### Traffic Retention

Continuous raw packet capture with point-in-time snapshots. Export PCAP files for offline analysis with Wireshark or other tools.

![Traffic Retention](https://github.com/kubeshark/assets/raw/master/png/snapshots.png)

[Snapshots guide →](https://docs.kubeshark.com/en/v2/traffic_snapshots)

---

## Features

| Feature | Description |
|---------|-------------|
| [**Raw Capture**](https://docs.kubeshark.com/en/v2/raw_capture) | Continuous cluster-wide packet capture with minimal overhead |
| [**Traffic Snapshots**](https://docs.kubeshark.com/en/v2/traffic_snapshots) | Point-in-time snapshots, export as PCAP for Wireshark |
| [**L7 API Dissection**](https://docs.kubeshark.com/en/v2/l7_api_dissection) | Request/response matching with full payloads and protocol parsing |
| [**Protocol Support**](https://docs.kubeshark.com/en/protocols) | HTTP, gRPC, GraphQL, Redis, Kafka, DNS, and more |
| [**TLS Decryption**](https://docs.kubeshark.com/en/encrypted_traffic) | eBPF-based decryption without key management |
| [**AI-Powered Analysis**](https://docs.kubeshark.com/en/v2/ai_powered_analysis) | Query cluster-wide network data with Claude, Cursor, or any MCP-compatible AI |
| [**Display Filters**](https://docs.kubeshark.com/en/v2/kfl2) | Wireshark-inspired display filters for precise traffic analysis |
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
