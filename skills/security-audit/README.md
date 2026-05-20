# Security Audit Skill

A Kubeshark MCP skill that teaches AI agents to perform systematic Kubernetes
network security audits using the MITRE ATT&CK framework. It examines DNS
queries, HTTP requests, L4 flows, and protocol-level payloads to detect
compromised workloads, C2 communication, data exfiltration, cryptomining,
lateral movement, and credential theft.

See [SKILL.md](SKILL.md) for the full methodology.

## Demo

The demo below shows a real security audit session against a compromised
`k8s-mule` namespace containing 21 workloads, 6 of which were actively
compromised with C2, cryptomining, secret theft, S3 exfiltration, port
scanning, and Redis reconnaissance.

### Claude Code Session

<!-- TODO: replace with animated GIF once recorded -->
![Security Audit Demo](https://raw.githubusercontent.com/kubeshark/assets/master/png/security-audit-demo.gif)

### Sample Audit Report

<!-- TODO: replace with animated GIF once recorded -->
![Security Audit Report](https://raw.githubusercontent.com/kubeshark/assets/master/png/security-audit-report.gif)
