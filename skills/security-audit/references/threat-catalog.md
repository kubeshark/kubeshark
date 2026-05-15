# Network Threat Catalog

22 network-observable threat patterns organized by MITRE ATT&CK tactic.
Each entry describes the attack, what it looks like on the wire, and how
to detect it with Kubeshark.

## Command & Control (TA0011)

### DGA Beaconing (T1568.002)
- **What**: Malware generates pseudo-random domain names daily and queries DNS
  for each. The C2 operator registers a few; most resolve to NXDOMAIN.
- **Wire signature**: Burst of DNS queries for high-entropy .com/.net domains
  with >80% NXDOMAIN response rate.
- **KFL**: `dns && dns_response && size(dns_answers) == 0` — then check for entropy in queried names.
- **Difficulty**: Medium. NXDOMAIN flood is distinctive but low-rate DGA can
  blend with legitimate DNS failures.

### HTTP C2 Beaconing (T1071.001)
- **What**: Implant calls home via HTTP GET at regular intervals, receiving
  tasking in the response body. Cobalt Strike, Meterpreter pattern.
- **Wire signature**: Periodic HTTP GET to fixed external URL at suspiciously
  regular intervals (30-60s). Outdated User-Agent (Mozilla/4.0). Session
  identifiers in URL path.
- **KFL**: `http && dst.name.contains("attacker")` or check for User-Agent anomalies.
- **Difficulty**: Medium. Regularity is the key anomaly.

### Encrypted C2 (T1573.002)
- **What**: C2 over HTTPS. Content is encrypted but TLS SNI reveals suspicious
  domain names.
- **Wire signature**: Outbound TLS to non-standard domains (darknet, cdn-mirror).
  DNS queries preceding the connection reveal the target.
- **KFL**: `dns && (dns_questions.exists(q, q.contains("darknet")) || dns_questions.exists(q, q.contains("cdn-mirror")))`.
- **Difficulty**: Hard. Encrypted, uses standard port 443.

### DNS-over-HTTPS C2 (T1572)
- **What**: Bypasses cluster DNS by sending queries as HTTPS to public DoH
  resolvers (cloudflare-dns.com, dns.google). C2 commands embedded in TXT
  responses.
- **Wire signature**: HTTP requests to DoH endpoints with `accept: application/dns-json`
  header. No corresponding queries on port 53.
- **KFL**: `http && (dst.name.contains("cloudflare-dns") || dst.name.contains("dns.google"))`.
- **Difficulty**: Hard. Looks like regular HTTPS to trusted providers.

## Exfiltration (TA0010)

### DNS Tunneling (T1048.003)
- **What**: Full bidirectional data channel over DNS using tools like iodine,
  dnscat2. Data encoded in long subdomain labels.
- **Wire signature**: High-frequency DNS queries (20+/burst) with subdomain
  labels near 63-byte limit. Mix of A, TXT, NULL query types.
- **KFL**: `dns && dns_questions.exists(q, q.contains("data-relay"))` or look for
  high query rates per source.
- **Difficulty**: Medium. Volume and long subdomains are distinctive.

### HTTP Header Exfiltration (T1048.001)
- **What**: Data exfiltrated in HTTP headers (Cookie, X-Trace-ID) disguised
  as analytics tracking. Low volume to evade detection.
- **Wire signature**: HTTP GET to analytics-looking URL with oversized Cookie
  or custom headers containing base64-encoded data.
- **KFL**: `http && dst.name.contains("cdn-provider")`.
- **Difficulty**: Hard. Low volume, standard HTTP, looks like analytics.

### DNS Credential Exfiltration (T1048.003)
- **What**: Stolen JWT tokens or credentials encoded in DNS TXT queries to
  attacker-controlled authoritative nameserver.
- **Wire signature**: DNS TXT queries with structured multi-label subdomains
  containing base64-like encoded data.
- **KFL**: `dns && dns_questions.exists(q, q.contains("steal-creds"))`.
- **Difficulty**: Medium. Multi-label structure is distinctive.

### gRPC Stream Exfiltration (T1048.001)
- **What**: Data exfiltration via gRPC (HTTP/2) POST to external endpoint.
  Blends with normal microservice traffic.
- **Wire signature**: HTTP/2 POST with `Content-Type: application/grpc` to
  external destination with exfil-related method names.
- **KFL**: `grpc && dst.name.contains("attacker")`.
- **Difficulty**: Hard. gRPC is normal in K8s. External destination is the signal.

## Lateral Movement (TA0008)

### K8s API Enumeration (T1613)
- **What**: Compromised pod uses mounted service account token to enumerate
  secrets, pods, RBAC bindings across all namespaces.
- **Wire signature**: HTTPS to kubernetes.default.svc with broad GET requests
  across /api/v1/secrets, /pods, /configmaps, /clusterrolebindings.
- **KFL**: `http && dst.port == 443 && path.contains("/api/v1/secrets")`.
- **Difficulty**: Medium. The fanout across resource types is the anomaly.

### SSRF to Internal Services (T1090)
- **What**: Pod probes cross-namespace internal services it shouldn't talk to —
  kube-dns metrics, Prometheus, Grafana, dashboards.
- **Wire signature**: HTTP to multiple ClusterIP services across namespaces
  from a single source pod.
- **KFL**: `http && src.pod.namespace == "k8s-mule" && dst.pod.namespace != "k8s-mule"`.
- **Difficulty**: Medium. Cross-namespace breadth is the signal.

### Port Scanning (T1046)
- **What**: Sweep of common ports across pod CIDR after initial access.
- **Wire signature**: Rapid TCP SYN from single source to many IPs on ports
  80, 443, 3306, 5432, 6379, 8080, 9090, 27017. High RST/timeout rate.
- **KFL**: `tcp && src.name == "network-diagnostics"`.
- **Difficulty**: Easy. Classic scan pattern — high fan-out, high failure rate.

### Service Fingerprinting (T1046)
- **What**: HTTP probes to discovery paths across multiple services to identify
  running software.
- **Wire signature**: HTTP GET to /version, /healthz, /.env, /actuator/info,
  /server-info. HEAD and OPTIONS methods. Multiple targets from one source.
- **KFL**: `http && (path == "/.env" || path == "/actuator/info")`.
- **Difficulty**: Medium. Path patterns are distinctive.

## Credential Access (TA0006)

### IMDS Metadata Theft (T1552.005)
- **What**: Query AWS/GCP instance metadata to steal IAM role credentials.
  The Capital One breach vector.
- **Wire signature**: HTTP to 169.254.169.254 with paths /latest/meta-data/iam/,
  /latest/user-data, /latest/api/token (PUT for IMDSv2).
- **KFL**: `dst.ip == "169.254.169.254"`.
- **Difficulty**: Easy. Destination IP is unique and unmistakable.

### Cloud API Abuse (T1078.004)
- **What**: Direct calls to AWS APIs (STS, S3, EC2) with stolen credentials
  from a workload pod.
- **Wire signature**: DNS for sts.amazonaws.com, s3.amazonaws.com. HTTPS
  requests with AWS Signature V4 Authorization headers.
- **KFL**: `dns && dns_questions.exists(q, q.contains("amazonaws.com"))`.
- **Difficulty**: Medium. Cloud API DNS from a non-controller pod is suspicious.

## Resource Hijacking (TA0040)

### Stratum Mining Protocol (T1496)
- **What**: XMRig/miner connecting to mining pool via Stratum JSON-RPC over TCP.
- **Wire signature**: TCP connection to port 3333/14433/45700 with JSON-RPC
  messages: mining.subscribe, mining.authorize, mining.submit.
- **KFL**: `dst.port == 3333`.
- **Difficulty**: Medium. Port 3333 is a well-known mining indicator.

### Mining Pool DNS (T1496)
- **What**: DNS resolution of known mining pool domains before connecting.
- **Wire signature**: DNS queries for domains containing minexmr, nanopool,
  mining-pool, hashvault, supportxmr.
- **KFL**: `dns && (dns_questions.exists(q, q.contains("minexmr")) || dns_questions.exists(q, q.contains("mining-pool")))`.
- **Difficulty**: Easy. Mining domain names are unmistakable.

### WebSocket Mining (T1496)
- **What**: Browser-based miner communicating via WebSocket on standard ports.
- **Wire signature**: HTTP Upgrade: websocket request to external host with
  mining-related URL path (/proxy?coin=, ?algo=randomx).
- **KFL**: `http && map_get(request.headers, "upgrade", "") == "websocket"`.
- **Difficulty**: Hard. WebSocket on port 80/443 looks normal. Only URL reveals intent.

## Protocol Abuse

### SQL Injection via PG Wire (T1190)
- **What**: SQL injection payloads sent through PostgreSQL wire protocol.
- **Wire signature**: PG protocol carrying UNION SELECT, information_schema,
  pg_shadow queries.
- **KFL**: `postgresql`.
- **Difficulty**: Medium. PG dissection reveals the SQL content directly.

### Redis Unauthorized Access (T1190)
- **What**: Unauthenticated Redis instance probed with dangerous commands.
- **Wire signature**: Redis protocol: CONFIG GET *, KEYS *, CLIENT LIST, DEBUG.
- **KFL**: `redis`.
- **Difficulty**: Easy. Redis command names are directly visible.

### Database Destruction (T1485)
- **What**: Ransomware pattern — SELECT * (data theft) then TRUNCATE/DROP (destruction).
- **Wire signature**: PG protocol showing SELECT followed by TRUNCATE on same table.
- **KFL**: `postgresql`.
- **Difficulty**: Medium. DDL commands in PG protocol are visible with dissection.

## Reconnaissance (TA0043)

### DNS Zone Enumeration (T1018)
- **What**: Brute-force DNS queries across namespaces to discover services.
  Includes SRV lookups and AXFR zone transfer attempts.
- **Wire signature**: High volume of DNS queries for *.svc.cluster.local patterns
  across many namespaces. Many NXDOMAIN responses.
- **KFL**: `dns && src.name == "service-discovery"`.
- **Difficulty**: Easy. Volume and cross-namespace pattern is obvious.

### gRPC Reflection Enumeration (T1046)
- **What**: Probing gRPC server reflection to discover API surfaces without
  needing proto files.
- **Wire signature**: HTTP/2 POST to /grpc.reflection.v1alpha.ServerReflection/
  ServerReflectionInfo across multiple services.
- **KFL**: `grpc && grpc_method.contains("Reflection")` or `http && path.contains("grpc.reflection")`.
- **Difficulty**: Medium. Reflection path is a known enumeration vector.
