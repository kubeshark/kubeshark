# L4 Flows Viewer

A simple single-page application to view Kubeshark L4 network flows.

## Prerequisites

- Kubeshark running with `/mcp/flows` endpoint enabled
- Modern web browser

## Usage

1. Start Kubeshark:
   ```bash
   kubeshark tap
   ```

2. Open `index.html` in your browser:
   ```bash
   open index.html
   # or
   python3 -m http.server 8080
   # then visit http://localhost:8080
   ```

3. Configure the Kubeshark URL (default: `http://localhost:8899`)

4. Click "Fetch Flows" or enable "Auto Refresh"

## Features

- **Filters**: Protocol, namespace, pod, service, IP
- **Aggregation**: Group by pod, namespace, node, or service
- **Stats**: Total flows, TCP/UDP counts, total bytes
- **RTT Metrics**: TCP handshake latency (p50/p90)
- **Auto Refresh**: Updates every 5 seconds

## API Endpoint

This viewer uses the `/mcp/flows` endpoint:

```
GET /mcp/flows?format=compact&limit=50&l4proto=tcp&ns=default
```

### Query Parameters

| Parameter | Description |
|-----------|-------------|
| `format` | `compact`, `full`, or `raw` |
| `l4proto` | `tcp` or `udp` |
| `ns` | Namespace filter |
| `pod` | Pod name filter |
| `svc` | Service name filter |
| `ip` | IP address filter |
| `limit` | Max results |
| `aggregate` | `pod`, `namespace`, `node`, `service` |

## Screenshot

```
+------------------------------------------------------------------+
| Kubeshark L4 Flows                                               |
+------------------------------------------------------------------+
| [URL: http://localhost:8899] [Proto: All] [NS:    ] [Fetch]     |
+------------------------------------------------------------------+
| 127 Flows | 98 TCP | 29 UDP | 45.2 MB                           |
+------------------------------------------------------------------+
| Proto | Client              | → | Server              | Bytes    |
|-------|---------------------|---|---------------------|----------|
| TCP   | frontend-abc        | → | api-svc:8080       | 1.2 MB   |
| TCP   | api-xyz             | → | postgres:5432      | 456 KB   |
| UDP   | coredns             | → | 8.8.8.8:53         | 12 KB    |
+------------------------------------------------------------------+
```
