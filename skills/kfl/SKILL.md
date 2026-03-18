---
name: kfl
description: >
  KFL2 (Kubeshark Filter Language) expert. Use this skill whenever the user needs to
  write, debug, or optimize KFL filters for Kubeshark traffic queries. Trigger on any
  mention of KFL, CEL filters, traffic filtering, display filters, query syntax,
  filter expressions, "how do I filter", "show me only", "find traffic where",
  protocol-specific queries (HTTP status codes, DNS lookups, Redis commands, Kafka topics),
  Kubernetes-aware filtering (by namespace, pod, service, label, annotation),
  L4 connection/flow filters, capture source filters, time-based queries, or any
  request to slice/search/narrow network traffic in Kubeshark. Also trigger when other
  skills need help constructing filters — KFL is the query language for all Kubeshark
  traffic analysis.
---

# KFL2 — Kubeshark Filter Language

You are a KFL2 expert. KFL2 is built on Google's CEL (Common Expression Language)
and is the query language for all Kubeshark traffic analysis. It operates as a
**display filter** — it doesn't affect what's captured, only what you see.

Think of KFL the way you think of SQL for databases or Google search syntax for
the web. Kubeshark captures and indexes all cluster traffic; KFL is how you
search it.

For the complete variable and field reference, see `references/kfl2-reference.md`.

## Core Syntax

KFL expressions are boolean CEL expressions. An empty filter matches everything.

### Operators

| Category | Operators |
|----------|-----------|
| Comparison | `==`, `!=`, `<`, `<=`, `>`, `>=` |
| Logical | `&&`, `\|\|`, `!` |
| Arithmetic | `+`, `-`, `*`, `/`, `%` |
| Membership | `in` |
| Ternary | `condition ? true_val : false_val` |

### String Functions

```
str.contains(substring)      // Substring search
str.startsWith(prefix)       // Prefix match
str.endsWith(suffix)         // Suffix match
str.matches(regex)           // Regex match
size(str)                    // String length
```

### Collection Functions

```
size(collection)             // List/map/string length
key in map                   // Key existence
map[key]                     // Value access
map_get(map, key, default)   // Safe access with default
value in list                // List membership
```

### Time Functions

```
timestamp("2026-03-14T22:00:00Z")   // Parse ISO timestamp
duration("5m")                        // Parse duration
now()                                 // Current time (snapshot at filter creation)
```

### Negation

```
!http                                // Everything that is NOT HTTP
http && status_code != 200           // HTTP responses that aren't 200
http && !path.contains("/health")    // Exclude health checks
!(src.pod.namespace == "kube-system")  // Exclude system namespace
```

## Protocol Detection

Boolean flags that indicate which protocol was detected. Use these as the first
filter term — they're fast and narrow the search space immediately.

| Flag | Protocol | Flag | Protocol |
|------|----------|------|----------|
| `http` | HTTP/1.1, HTTP/2 | `redis` | Redis |
| `dns` | DNS | `kafka` | Kafka |
| `tls` | TLS/SSL | `amqp` | AMQP |
| `tcp` | TCP | `ldap` | LDAP |
| `udp` | UDP | `ws` | WebSocket |
| `sctp` | SCTP | `gql` | GraphQL (v1+v2) |
| `icmp` | ICMP | `gqlv1` / `gqlv2` | GraphQL version-specific |
| `radius` | RADIUS | `conn` / `flow` | L4 connection/flow tracking |
| `diameter` | Diameter | `tcp_conn` / `udp_conn` | Transport-specific connections |

## Kubernetes Context

The most common starting point. Filter by where traffic originates or terminates.

### Pod and Service Fields

```
src.pod.name == "orders-594487879c-7ddxf"
dst.pod.namespace == "production"
src.service.name == "api-gateway"
dst.service.namespace == "payments"
```

Pod fields fall back to service data when pod info is unavailable, so
`dst.pod.namespace` works even for service-level entries.

### Aggregate Collections

Match against any direction (src or dst):

```
"production" in namespaces           // Any namespace match
"orders" in pods                     // Any pod name match
"api-gateway" in services            // Any service name match
```

### Labels and Annotations

```
map_get(local_labels, "app", "") == "checkout"   // Safe access with default
map_get(remote_labels, "version", "") == "canary"
"tier" in local_labels                            // Label existence check
```

Always use `map_get()` for labels and annotations — direct access like
`local_labels["app"]` errors if the key doesn't exist.

### Node and Process

```
node_name == "ip-10-0-25-170.ec2.internal"
local_process_name == "nginx"
remote_process_name.contains("postgres")
```

### DNS Resolution

```
src.dns == "api.example.com"
dst.dns.contains("redis")
```

## HTTP Filtering

HTTP is the most common protocol for API-level investigation.

### Fields

| Field | Type | Example |
|-------|------|---------|
| `method` | string | `"GET"`, `"POST"`, `"PUT"`, `"DELETE"` |
| `url` | string | Full path + query: `"/api/users?id=123"` |
| `path` | string | Path only: `"/api/users"` |
| `status_code` | int | `200`, `404`, `500` |
| `http_version` | string | `"HTTP/1.1"`, `"HTTP/2"` |
| `request.headers` | map | `request.headers["content-type"]` |
| `response.headers` | map | `response.headers["server"]` |
| `request.cookies` | map | `request.cookies["session"]` |
| `response.cookies` | map | `response.cookies["token"]` |
| `query_string` | map | `query_string["id"]` |
| `request_body_size` | int | Request body bytes |
| `response_body_size` | int | Response body bytes |
| `elapsed_time` | int | Duration in **microseconds** |

### Common Patterns

```
// Error investigation
http && status_code >= 500                           // Server errors
http && status_code == 429                           // Rate limiting
http && status_code >= 400 && status_code < 500      // Client errors

// Endpoint targeting
http && method == "POST" && path.contains("/orders")
http && url.matches(".*/api/v[0-9]+/users.*")

// Performance
http && elapsed_time > 5000000                       // > 5 seconds
http && response_body_size > 1000000                 // > 1MB responses

// Header inspection
http && "authorization" in request.headers
http && request.headers["content-type"] == "application/json"

// GraphQL (subset of HTTP)
gql && method == "POST" && status_code >= 400
```

## DNS Filtering

DNS issues are often the hidden root cause of outages.

| Field | Type | Description |
|-------|------|-------------|
| `dns_questions` | []string | Question domain names |
| `dns_answers` | []string | Answer domain names |
| `dns_question_types` | []string | Record types: A, AAAA, CNAME, MX, TXT, SRV, PTR |
| `dns_request` | bool | Is request |
| `dns_response` | bool | Is response |
| `dns_request_length` | int | Request size |
| `dns_response_length` | int | Response size |

```
dns && "api.external-service.com" in dns_questions
dns && dns_response && status_code != 0              // Failed lookups
dns && "A" in dns_question_types                     // A record queries
dns && size(dns_questions) > 1                       // Multi-question
```

## Database and Messaging Protocols

### Redis

```
redis && redis_type == "GET"                         // Command type
redis && redis_key.startsWith("session:")            // Key pattern
redis && redis_command.contains("DEL")               // Command search
redis && redis_total_size > 10000                    // Large operations
```

### Kafka

```
kafka && kafka_api_key_name == "PRODUCE"             // Produce operations
kafka && kafka_client_id == "payment-processor"      // Client filtering
kafka && kafka_request_summary.contains("orders")    // Topic filtering
kafka && kafka_size > 10000                          // Large messages
```

### AMQP, LDAP, RADIUS, Diameter

```
amqp && amqp_method == "basic.publish"               // AMQP publish
ldap && ldap_type == "bind"                          // LDAP bind requests
radius && radius_code_name == "Access-Request"       // RADIUS auth
diameter && diameter_method.contains("Credit")       // Diameter credit control
```

For the full variable list for these protocols, see `references/kfl2-reference.md`.

## Transport Layer (L4)

### TCP/UDP Fields

```
tcp && tcp_error_type != ""                          // TCP errors
udp && udp_length > 1000                             // Large UDP packets
```

### Connection Tracking

```
conn && conn_state == "open"                         // Active connections
conn && conn_local_bytes > 1000000                   // High-volume
conn && "HTTP" in conn_l7_detected                   // L7 protocol detection
tcp_conn && conn_state == "closed"                   // Closed TCP connections
```

### Flow Tracking (with Rate Metrics)

```
flow && flow_local_pps > 1000                        // High packet rate
flow && flow_local_bps > 1000000                     // High bandwidth
flow && flow_state == "closed" && "TLS" in flow_l7_detected
tcp_flow && flow_local_bps > 5000000                 // High-throughput TCP
```

## Network Layer

```
src.ip == "10.0.53.101"
dst.ip.startsWith("192.168.")
src.port == 8080
dst.port >= 8000 && dst.port <= 9000
```

## Time-Based Filtering

```
timestamp > timestamp("2026-03-14T22:00:00Z")
timestamp >= timestamp("2026-03-14T22:00:00Z") && timestamp <= timestamp("2026-03-14T23:00:00Z")
timestamp > now() - duration("5m")                   // Last 5 minutes
elapsed_time > 2000000                               // Older than 2 seconds
```

## Building Filters: Progressive Narrowing

The most effective investigation technique — start broad, add constraints:

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

## Performance Tips

1. **Protocol flags first** — `http && ...` is faster than `... && http`
2. **`startsWith`/`endsWith` over `contains`** — prefix/suffix checks are faster
3. **Specific ports before string ops** — `dst.port == 80` is cheaper than `url.contains(...)`
4. **Use `map_get` for labels** — avoids errors on missing keys
5. **Keep filters simple** — CEL short-circuits on `&&`, so put cheap checks first

## Type Safety

KFL2 is statically typed. Common gotchas:

- `status_code` is `int`, not string — use `status_code == 200`, not `"200"`
- `elapsed_time` is in **microseconds** — 5 seconds = `5000000`
- `timestamp` requires `timestamp()` function — not a raw string
- Map access on missing keys errors — use `key in map` or `map_get()` first
- List membership uses `value in list` — not `list.contains(value)`
