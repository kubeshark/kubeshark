# KFL2 Complete Variable and Field Reference

This is the exhaustive reference for every variable available in KFL2 filters.
KFL2 is built on Google's CEL (Common Expression Language) and evaluates against
Kubeshark's protobuf-based `BaseEntry` structure.

## Most Commonly Used Variables

These are the variables you'll reach for in 90% of investigations:

| Variable | Type | What it's for |
|----------|------|---------------|
| `status_code` | int | HTTP response status (200, 404, 500) |
| `method` | string | HTTP method (GET, POST, PUT, DELETE) |
| `path` | string | URL path without query string |
| `dst.pod.namespace` | string | Where traffic is going (namespace) |
| `dst.service.name` | string | Where traffic is going (service) |
| `src.pod.name` | string | Where traffic comes from (pod) |
| `elapsed_time` | int | Request duration in microseconds |
| `dns_questions` | []string | DNS domains being queried |
| `namespaces` | []string | All namespaces involved (src + dst) |

## Network-Level Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `src.ip` | string | Source IP address | `"10.0.53.101"` |
| `dst.ip` | string | Destination IP address | `"192.168.1.1"` |
| `src.port` | int | Source port number | `43210` |
| `dst.port` | int | Destination port number | `8080` |
| `protocol` | string | Detected protocol type | `"HTTP"`, `"DNS"` |

## Identity and Metadata Variables

| Variable | Type | Description |
|----------|------|-------------|
| `id` | int | BaseEntry unique identifier (assigned by sniffer) |
| `node_id` | string | Node identifier (assigned by hub) |
| `index` | int | Entry index for stream uniqueness |
| `stream` | string | Stream identifier (hex string) |
| `timestamp` | timestamp | Event time (UTC), use with `timestamp()` function |
| `elapsed_time` | int | Age since timestamp in microseconds |
| `worker` | string | Worker identifier |

## Cross-Reference Variables

| Variable | Type | Description |
|----------|------|-------------|
| `conn_id` | int | L7 to L4 connection cross-reference ID |
| `flow_id` | int | L7 to L4 flow cross-reference ID |
| `has_pcap` | bool | Whether PCAP data is available for this entry |

## Capture Source Variables

| Variable | Type | Description | Values |
|----------|------|-------------|--------|
| `capture_source` | string | Canonical capture source | `"unspecified"`, `"af_packet"`, `"ebpf"`, `"ebpf_tls"` |
| `capture_backend` | string | Backend family | `"af_packet"`, `"ebpf"` |
| `capture_source_code` | int | Numeric enum | 0=unspecified, 1=af_packet, 2=ebpf, 3=ebpf_tls |
| `capture` | map | Nested map access | `capture["source"]`, `capture["backend"]` |

## Protocol Detection Flags

Boolean variables indicating detected protocol. Use as first filter term for performance.

| Variable | Protocol | Variable | Protocol |
|----------|----------|----------|----------|
| `http` | HTTP/1.1, HTTP/2 | `redis` | Redis |
| `dns` | DNS | `kafka` | Kafka |
| `tls` | TLS/SSL handshake | `amqp` | AMQP messaging |
| `tcp` | TCP transport | `ldap` | LDAP directory |
| `udp` | UDP transport | `ws` | WebSocket |
| `sctp` | SCTP streaming | `gql` | GraphQL (v1 or v2) |
| `icmp` | ICMP | `gqlv1` | GraphQL v1 only |
| `radius` | RADIUS auth | `gqlv2` | GraphQL v2 only |
| `diameter` | Diameter | `conn` | L4 connection tracking |
| `flow` | L4 flow tracking | `tcp_conn` | TCP connection tracking |
| `tcp_flow` | TCP flow tracking | `udp_conn` | UDP connection tracking |
| `udp_flow` | UDP flow tracking | | |

## HTTP Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `method` | string | HTTP method | `"GET"`, `"POST"`, `"PUT"`, `"DELETE"`, `"PATCH"` |
| `url` | string | Full URL path and query string | `"/api/users?id=123"` |
| `path` | string | URL path component (no query) | `"/api/users"` |
| `status_code` | int | HTTP response status code | `200`, `404`, `500` |
| `http_version` | string | HTTP protocol version | `"HTTP/1.1"`, `"HTTP/2"` |
| `query_string` | map[string]string | Parsed URL query parameters | `query_string["id"]` → `"123"` |
| `request.headers` | map[string]string | Request HTTP headers | `request.headers["content-type"]` |
| `response.headers` | map[string]string | Response HTTP headers | `response.headers["server"]` |
| `request.cookies` | map[string]string | Request cookies | `request.cookies["session"]` |
| `response.cookies` | map[string]string | Response cookies | `response.cookies["token"]` |
| `request_headers_size` | int | Request headers size in bytes | |
| `request_body_size` | int | Request body size in bytes | |
| `response_headers_size` | int | Response headers size in bytes | |
| `response_body_size` | int | Response body size in bytes | |

GraphQL requests have `gql` (or `gqlv1`/`gqlv2`) set to true and all HTTP
variables available.

**Example**: `http && method == "POST" && status_code >= 500 && path.contains("/api")`

## DNS Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `dns_questions` | []string | Question domain names (request + response) | `["example.com"]` |
| `dns_answers` | []string | Answer domain names | `["1.2.3.4"]` |
| `dns_question_types` | []string | Record types in questions | `["A"]`, `["AAAA"]`, `["CNAME"]` |
| `dns_request` | bool | Is DNS request message | |
| `dns_response` | bool | Is DNS response message | |
| `dns_request_length` | int | DNS request size in bytes (0 if absent) | |
| `dns_response_length` | int | DNS response size in bytes (0 if absent) | |
| `dns_total_size` | int | Sum of request + response sizes | |

Supported question types: A, AAAA, NS, CNAME, SOA, MX, TXT, SRV, PTR, ANY.

**Example**: `dns && dns_response && status_code != 0` (failed DNS lookups)

## TLS Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `tls` | bool | TLS payload detected | |
| `tls_summary` | string | TLS handshake summary | `"ClientHello"`, `"ServerHello"` |
| `tls_info` | string | TLS connection details | `"TLS 1.3, AES-256-GCM"` |
| `tls_request_size` | int | TLS request size in bytes | |
| `tls_response_size` | int | TLS response size in bytes | |
| `tls_total_size` | int | Sum of request + response (computed if not provided) | |

## TCP Variables

| Variable | Type | Description |
|----------|------|-------------|
| `tcp` | bool | TCP payload detected |
| `tcp_method` | string | TCP method information |
| `tcp_payload` | bytes | Raw TCP payload data |
| `tcp_error_type` | string | TCP error type (empty if none) |
| `tcp_error_message` | string | TCP error message (empty if none) |

## UDP Variables

| Variable | Type | Description |
|----------|------|-------------|
| `udp` | bool | UDP payload detected |
| `udp_length` | int | UDP packet length |
| `udp_checksum` | int | UDP checksum value |
| `udp_payload` | bytes | Raw UDP payload data |

## SCTP Variables

| Variable | Type | Description |
|----------|------|-------------|
| `sctp` | bool | SCTP payload detected |
| `sctp_checksum` | int | SCTP checksum value |
| `sctp_chunk_type` | string | SCTP chunk type |
| `sctp_length` | int | SCTP chunk length |

## ICMP Variables

| Variable | Type | Description |
|----------|------|-------------|
| `icmp` | bool | ICMP payload detected |
| `icmp_type` | string | ICMP type code |
| `icmp_version` | int | ICMP version (4 or 6) |
| `icmp_length` | int | ICMP message length |

## WebSocket Variables

| Variable | Type | Description | Values |
|----------|------|-------------|--------|
| `ws` | bool | WebSocket payload detected | |
| `ws_opcode` | string | WebSocket operation code | `"text"`, `"binary"`, `"close"`, `"ping"`, `"pong"` |
| `ws_request` | bool | Is WebSocket request | |
| `ws_response` | bool | Is WebSocket response | |
| `ws_request_payload_data` | string | Request payload (safely truncated) | |
| `ws_request_payload_length` | int | Request payload length in bytes | |
| `ws_response_payload_length` | int | Response payload length in bytes | |

## Redis Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `redis` | bool | Redis payload detected | |
| `redis_type` | string | Redis command verb | `"GET"`, `"SET"`, `"DEL"`, `"HGET"` |
| `redis_command` | string | Full Redis command line | `"GET session:1234"` |
| `redis_key` | string | Key (truncated to 64 bytes) | `"session:1234"` |
| `redis_request_size` | int | Request size (0 if absent) | |
| `redis_response_size` | int | Response size (0 if absent) | |
| `redis_total_size` | int | Sum of request + response | |

**Example**: `redis && redis_type == "GET" && redis_key.startsWith("session:")`

## Kafka Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `kafka` | bool | Kafka payload detected | |
| `kafka_api_key` | int | Kafka API key number | 0=FETCH, 1=PRODUCE |
| `kafka_api_key_name` | string | Human-readable API operation | `"PRODUCE"`, `"FETCH"` |
| `kafka_client_id` | string | Kafka client identifier | `"payment-processor"` |
| `kafka_size` | int | Message size (request preferred, else response) | |
| `kafka_request` | bool | Is Kafka request | |
| `kafka_response` | bool | Is Kafka response | |
| `kafka_request_summary` | string | Request summary/topic | `"orders-topic"` |
| `kafka_request_size` | int | Request size (0 if absent) | |
| `kafka_response_size` | int | Response size (0 if absent) | |

**Example**: `kafka && kafka_api_key_name == "PRODUCE" && kafka_request_summary.contains("orders")`

## AMQP Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `amqp` | bool | AMQP payload detected | |
| `amqp_method` | string | AMQP method name | `"basic.publish"`, `"channel.open"` |
| `amqp_summary` | string | Operation summary | |
| `amqp_request` | bool | Is AMQP request | |
| `amqp_response` | bool | Is AMQP response | |
| `amqp_request_length` | int | Request length (0 if absent) | |
| `amqp_response_length` | int | Response length (0 if absent) | |
| `amqp_total_size` | int | Sum of request + response | |

## LDAP Variables

| Variable | Type | Description |
|----------|------|-------------|
| `ldap` | bool | LDAP payload detected |
| `ldap_type` | string | LDAP operation type (request preferred) |
| `ldap_summary` | string | Operation summary |
| `ldap_request` | bool | Is LDAP request |
| `ldap_response` | bool | Is LDAP response |
| `ldap_request_length` | int | Request length (0 if absent) |
| `ldap_response_length` | int | Response length (0 if absent) |
| `ldap_total_size` | int | Sum of request + response |

## RADIUS Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `radius` | bool | RADIUS payload detected | |
| `radius_code` | int | RADIUS code (request preferred) | |
| `radius_code_name` | string | Code name | `"Access-Request"` |
| `radius_request` | bool | Is RADIUS request | |
| `radius_response` | bool | Is RADIUS response | |
| `radius_request_authenticator` | string | Request authenticator (hex) | |
| `radius_request_length` | int | Request size (0 if absent) | |
| `radius_response_length` | int | Response size (0 if absent) | |
| `radius_total_size` | int | Sum of request + response | |

## Diameter Variables

| Variable | Type | Description |
|----------|------|-------------|
| `diameter` | bool | Diameter payload detected |
| `diameter_method` | string | Method name (request preferred) |
| `diameter_summary` | string | Operation summary |
| `diameter_request` | bool | Is Diameter request |
| `diameter_response` | bool | Is Diameter response |
| `diameter_request_length` | int | Request size (0 if absent) |
| `diameter_response_length` | int | Response size (0 if absent) |
| `diameter_total_size` | int | Sum of request + response |

## L4 Connection Tracking Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `conn` | bool | Connection tracking entry | |
| `conn_state` | string | Connection state | `"open"`, `"in_progress"`, `"closed"` |
| `conn_local_pkts` | int | Packets from local peer | |
| `conn_local_bytes` | int | Bytes from local peer | |
| `conn_remote_pkts` | int | Packets from remote peer | |
| `conn_remote_bytes` | int | Bytes from remote peer | |
| `conn_l7_detected` | []string | L7 protocols detected on connection | `["HTTP", "TLS"]` |
| `conn_group_id` | int | Connection group identifier | |

**Example**: `conn && conn_state == "open" && conn_local_bytes > 1000000` (high-volume open connections)

## L4 Flow Tracking Variables

Flows extend connections with rate metrics (packets/bytes per second).

| Variable | Type | Description |
|----------|------|-------------|
| `flow` | bool | Flow tracking entry |
| `flow_state` | string | Flow state (`"open"`, `"in_progress"`, `"closed"`) |
| `flow_local_pkts` | int | Packets from local peer |
| `flow_local_bytes` | int | Bytes from local peer |
| `flow_remote_pkts` | int | Packets from remote peer |
| `flow_remote_bytes` | int | Bytes from remote peer |
| `flow_local_pps` | int | Local packets per second |
| `flow_local_bps` | int | Local bytes per second |
| `flow_remote_pps` | int | Remote packets per second |
| `flow_remote_bps` | int | Remote bytes per second |
| `flow_l7_detected` | []string | L7 protocols detected on flow |
| `flow_group_id` | int | Flow group identifier |

**Example**: `tcp_flow && flow_local_bps > 5000000` (high-bandwidth TCP flows)

## Kubernetes Variables

### Pod and Service (Directional)

| Variable | Type | Description |
|----------|------|-------------|
| `src.pod.name` | string | Source pod name |
| `src.pod.namespace` | string | Source pod namespace |
| `dst.pod.name` | string | Destination pod name |
| `dst.pod.namespace` | string | Destination pod namespace |
| `src.service.name` | string | Source service name |
| `src.service.namespace` | string | Source service namespace |
| `dst.service.name` | string | Destination service name |
| `dst.service.namespace` | string | Destination service namespace |

**Fallback behavior**: Pod namespace/name fields automatically fall back to
service data when pod info is unavailable. This means `dst.pod.namespace` works
even when only service-level resolution exists.

**Example**: `src.service.name == "api-gateway" && dst.pod.namespace == "production"`

### Aggregate Collections (Non-Directional)

| Variable | Type | Description |
|----------|------|-------------|
| `namespaces` | []string | All namespaces (src + dst, pod + service) |
| `pods` | []string | All pod names (src + dst) |
| `services` | []string | All service names (src + dst) |

### Labels and Annotations

| Variable | Type | Description |
|----------|------|-------------|
| `local_labels` | map[string]string | Kubernetes labels of local peer |
| `local_annotations` | map[string]string | Kubernetes annotations of local peer |
| `remote_labels` | map[string]string | Kubernetes labels of remote peer |
| `remote_annotations` | map[string]string | Kubernetes annotations of remote peer |

Use `map_get(local_labels, "key", "default")` for safe access that won't error
on missing keys.

**Example**: `map_get(local_labels, "app", "") == "checkout" && "production" in namespaces`

### Node Information

| Variable | Type | Description |
|----------|------|-------------|
| `node` | map | Nested: `node["name"]`, `node["ip"]` |
| `node_name` | string | Node name (flat alias) |
| `node_ip` | string | Node IP (flat alias) |
| `local_node_name` | string | Node name of local peer |
| `remote_node_name` | string | Node name of remote peer |

### Process Information

| Variable | Type | Description |
|----------|------|-------------|
| `local_process_name` | string | Process name on local peer |
| `remote_process_name` | string | Process name on remote peer |

### DNS Resolution

| Variable | Type | Description |
|----------|------|-------------|
| `src.dns` | string | DNS resolution of source IP |
| `dst.dns` | string | DNS resolution of destination IP |
| `dns_resolutions` | []string | All DNS resolutions (deduplicated) |

### Resolution Status

| Variable | Type | Values |
|----------|------|--------|
| `local_resolution_status` | string | `""` (resolved), `"no_node_mapping"`, `"rpc_error"`, `"rpc_empty"`, `"cache_miss"`, `"queue_full"` |
| `remote_resolution_status` | string | Same as above |

## Default Values

When a variable is not present in an entry, KFL2 uses these defaults:

| Type | Default |
|------|---------|
| string | `""` |
| int | `0` |
| bool | `false` |
| list | `[]` |
| map | `{}` |
| bytes | `[]` |

## Protocol Variable Precedence

For protocols with request/response pairs (Kafka, RADIUS, Diameter), merged
fields prefer the **request** side. If no request exists, the response value
is used. Size totals are always computed as `request_size + response_size`.

## CEL Language Features

KFL2 supports the full CEL specification:

- **Short-circuit evaluation**: `&&` stops on first false, `||` stops on first true
- **Ternary**: `condition ? value_if_true : value_if_false`
- **Regex**: `str.matches("pattern")` uses RE2 syntax
- **Type coercion**: Timestamps require `timestamp()`, durations require `duration()`
- **Null safety**: Use `in` operator or `map_get()` before accessing map keys

For the full CEL specification, see the
[CEL Language Definition](https://github.com/google/cel-spec/blob/master/doc/langdef.md).
