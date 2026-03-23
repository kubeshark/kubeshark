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

---

### Network Traffic Indexing

Kubeshark indexes cluster-wide network traffic by parsing it according to protocol specifications, with support for HTTP, gRPC, Redis, Kafka, DNS, and more. This enables queries using Kubernetes semantics (e.g. pod, namespace, node), API semantics (e.g. path, headers, status), and network semantics (e.g. IP, port). No code instrumentation required.

![API context](https://github.com/kubeshark/assets/raw/master/png/api_context.png)

[Learn more →](https://docs.kubeshark.com/en/v2/l7_api_dissection)

### Workload Dependency Map

A visual map of how workloads communicate, showing dependencies, traffic volume, and protocol usage across the cluster.

![Service Map](https://github.com/kubeshark/assets/raw/master/png/servicemap.png)

[Learn more →](https://docs.kubeshark.com/en/v2/service_map)

### Traffic Retention & PCAP Export

Capture and retain raw network traffic cluster-wide. Download PCAPs scoped by time range, nodes, workloads, and IPs — ready for Wireshark or any PCAP-compatible tool.

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
