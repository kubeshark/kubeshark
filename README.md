<p align="center">
  <img src="https://raw.githubusercontent.com/kubeshark/assets/master/svg/kubeshark-logo.svg" alt="Kubeshark" height="120px"/>
</p>

<p align="center">
    <a href="https://github.com/kubeshark/kubeshark/releases/latest"><img alt="Release" src="https://img.shields.io/github/v/release/kubeshark/kubeshark?logo=GitHub&style=flat-square"></a>
    <a href="https://hub.docker.com/r/kubeshark/worker"><img alt="Docker pulls" src="https://img.shields.io/docker/pulls/kubeshark/worker?color=%23099cec&logo=Docker&style=flat-square"></a>
    <a href="https://discord.gg/WkvRGMUcx7"><img alt="Discord" src="https://img.shields.io/discord/1042559155224973352?logo=Discord&style=flat-square&label=discord"></a>
    <a href="https://join.slack.com/t/kubeshark/shared_invite/zt-3jdcdgxdv-1qNkhBh9c6CFoE7bSPkpBQ"><img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-green?logo=Slack&style=flat-square"></a>
</p>

<p align="center"><b>Network Intelligence for Kubernetes</b></p>

<p align="center">
  <a href="https://demo.kubeshark.com/">Live Demo</a> · <a href="https://docs.kubeshark.com">Docs</a>
</p>

---

**Cluster-wide, real-time visibility into every packet, API call, and service interaction.** Replay any moment in time. Resolve incidents at the speed of LLMs. 100% on-premises.

<!-- TODO: Hero image -->
![Kubeshark](https://github.com/kubeshark/assets/raw/master/png/stream.png)

---

## Get Started

```bash
helm repo add kubeshark https://helm.kubeshark.com
helm install kubeshark kubeshark/kubeshark
```

Dashboard opens automatically. You're capturing traffic.

**With AI** — connect your assistant and debug with natural language:

```bash
brew install kubeshark
claude mcp add kubeshark -- kubeshark mcp
```

> *"Why did checkout fail at 2:15 PM?"*
> *"Which services have error rates above 1%?"*

[MCP setup guide →](https://docs.kubeshark.com/en/mcp)

---

## Why Kubeshark

- **Instant root cause** — trace requests across services, see exact errors
- **Zero instrumentation** — no code changes, no SDKs, just deploy
- **Full payload capture** — request/response bodies, headers, timing
- **TLS decryption** — see encrypted traffic without managing keys
- **AI-ready** — query traffic with natural language via [MCP](https://docs.kubeshark.com/en/mcp)

---

### Traffic Analysis and API Dissection

Capture and inspect every API call across your cluster—HTTP, gRPC, Redis, Kafka, DNS, and more. Request/response matching with full payloads, parsed according to protocol specifications. Headers, timing, and complete context. Zero instrumentation required.

[Learn more →](https://docs.kubeshark.com/en/traffic)

### Service Map

Visualize how your services communicate. See dependencies, traffic flow, and identify anomalies at a glance.

<!-- TODO: Service map screenshot -->
![Service Map](https://github.com/kubeshark/assets/raw/master/png/servicemap.png)

[Learn more →](https://docs.kubeshark.com/en/service-map)

### AI Debugging

Connect your AI assistant and investigate incidents using natural language. Ask questions, get answers—no dashboards, no query languages.

> *"Why did checkout fail at 2:15 PM?"*
> *"Which services have error rates above 1%?"*
> *"Trace request abc123 through all services"*

[MCP setup guide →](https://docs.kubeshark.com/en/mcp)

### Traffic Retention

Retain every packet. Take snapshots. Export PCAP files. Replay any moment in time.

<!-- TODO: PCAP/Snapshot screenshot -->
![Traffic Retention](https://github.com/kubeshark/assets/raw/master/png/snapshots.png)

[Snapshots guide →](https://docs.kubeshark.com/en/pcap)

---

## Features

| Feature | Description |
|---------|-------------|
| [**Raw Capture**](https://docs.kubeshark.com/en/v2/raw_capture) | Continuous cluster-wide packet capture with minimal overhead |
| [**Traffic Snapshots**](https://docs.kubeshark.com/en/v2/traffic_snapshots) | Point-in-time snapshots, export as PCAP for Wireshark |
| [**L7 API Dissection**](https://docs.kubeshark.com/en/v2/l7_api_dissection) | Request/response matching with full payloads and protocol parsing |
| [**Protocol Support**](https://docs.kubeshark.com/en/protocols) | HTTP, gRPC, GraphQL, Redis, Kafka, DNS, and more |
| [**TLS Decryption**](https://docs.kubeshark.com/en/encrypted_traffic) | eBPF-based decryption without key management |
| [**AI-Powered Analysis**](https://docs.kubeshark.com/en/mcp) | Query traffic with Claude, Cursor, or any MCP-compatible AI |
| [**KFL Filtering**](https://docs.kubeshark.com/en/filtering) | Wireshark-inspired display filters for precise traffic analysis |
| [**Long-term Retention**](https://docs.kubeshark.com/en/long_term_retention) | Upload to S3/GCS for compliance and forensics |
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
