# KFL Quick Reference: Security Audit Filters

## DNS Threat Hunting
```
dns                                                    // All DNS traffic
dns && dns_response && size(dns_answers) == 0          // Failed lookups (NXDOMAIN — no answers)
dns && dns_questions.exists(q, q.contains("minexmr"))  // Mining pool DNS
dns && dns_questions.exists(q, q.contains("nanopool"))  // Mining pool DNS
dns && dns_questions.exists(q, q.contains("amazonaws")) // Cloud API resolution
dns && dns_questions.exists(q, q.contains("cloudflare-dns"))  // DoH bypass
dns && dns_questions.exists(q, q.contains("dns.google"))      // DoH bypass
```

## External Communication
```
http && dst.name.contains("attacker")                  // Known-bad destinations
http && map_get(request.headers, "user-agent", "").contains("Mozilla/4.0")  // Suspicious UA
http && map_get(request.headers, "accept", "").contains("dns-json")         // DoH requests
http && map_get(request.headers, "upgrade", "") == "websocket"              // WebSocket (potential mining)
```

## Lateral Movement
```
src.pod.namespace != dst.pod.namespace                 // Cross-namespace traffic
http && path.startsWith("/api/v1/secrets")             // Secret enumeration
http && path == "/.env"                                // Service fingerprinting
http && path == "/actuator/info"                       // Spring Boot fingerprinting
http && path == "/version"                             // Version fingerprinting
```

## Protocol Inspection
```
postgresql                                             // PostgreSQL wire protocol
postgresql && postgresql_query.contains("UNION SELECT") // SQL injection patterns
postgresql && !postgresql_success                       // Failed PostgreSQL queries
redis                                                  // Redis protocol
grpc                                                   // gRPC calls (native detection)
grpc && grpc_method.contains("Reflection")             // gRPC reflection enumeration
```

## Credential Theft
```
dst.ip == "169.254.169.254"                            // IMDS access
http && path.contains("/meta-data/iam")                // IAM credential paths
http && map_get(request.headers, "authorization", "").startsWith("AWS4-HMAC-SHA256")  // Stolen AWS creds
http && "x-aws-ec2-metadata-token-ttl-seconds" in request.headers                     // IMDSv2 token request
```

## Resource Hijacking
```
dst.port == 3333                                       // Stratum mining (standard)
dst.port == 14433                                      // Stratum mining (alt)
dst.port == 45700                                      // Stratum mining (alt)
dst.port == 4444                                       // Reverse shell / backdoor
```

## Per-Namespace Scoping

Add namespace filters to any query above:
```
dns && src.pod.namespace == "k8s-mule"                 // DNS from specific namespace
http && src.pod.namespace == "k8s-mule"                // HTTP from specific namespace
redis && src.pod.namespace == "k8s-mule"               // Redis from specific namespace
```
