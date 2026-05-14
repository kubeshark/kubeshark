# Kubeshark Helm Values Reference

Complete reference for all Kubeshark Helm chart values. Use this when building
custom `values.yaml` files or `--set` flags.

## Docker Images

```yaml
tap:
  docker:
    registry: docker.io/kubeshark    # Docker registry
    tag: ""                          # Image tag (empty = chart appVersion)
    tagLocked: true                  # Lock to specific tag
    imagePullPolicy: Always          # Always, IfNotPresent, Never
    imagePullSecrets: []             # Registry pull secrets
    overrideImage:                   # Override individual component images
      worker: ""
      hub: ""
      front: ""
    overrideTag:                     # Override individual component tags
      worker: ""
      hub: ""
      front: ""
```

## Proxy / Port-Forward

```yaml
tap:
  proxy:
    worker:
      srvPort: 48999
    hub:
      srvPort: 8898
    front:
      port: 8899                     # Local port for port-forward
    host: 127.0.0.1                  # Bind address
```

## Pod Targeting

```yaml
tap:
  regex: .*                          # Pod name regex filter
  namespaces: []                     # Target namespaces (empty = all)
  excludedNamespaces: []             # Namespaces to exclude
  bpfOverride: ""                    # Custom BPF filter override
```

## Capture & Dissection

```yaml
tap:
  capture:
    dissection:
      enabled: true                  # Enable L7 dissection
      stopAfter: 5m                  # Auto-stop dissection after duration
    captureSelf: false               # Capture Kubeshark's own traffic
    raw:
      enabled: true                  # Enable raw packet capture (needed for snapshots)
      storageSize: 1Gi               # FIFO buffer size per node
    dbMaxSize: 500Mi                 # Max L7 database size per node
  delayedDissection:
    cpu: "1"                         # CPU for delayed dissection jobs
    memory: 4Gi                      # Memory for delayed dissection jobs
    storageSize: ""                  # Storage for delayed dissection
    storageClass: ""                 # Storage class for delayed dissection
```

## Snapshots

```yaml
tap:
  snapshots:
    local:
      storageClass: ""               # Storage class for local snapshots
      storageSize: 20Gi              # PVC size for local snapshots
    cloud:
      provider: ""                   # s3, gcs, or azblob
      prefix: ""                     # Path prefix in bucket
      configMaps: []                 # Additional ConfigMaps to mount
      secrets: []                    # Additional Secrets to mount
      s3:
        bucket: ""
        region: ""
        accessKey: ""
        secretKey: ""
        roleArn: ""                  # IAM role ARN (IRSA)
        externalId: ""               # STS external ID
      azblob:
        storageAccount: ""
        container: ""
        storageKey: ""
      gcs:
        bucket: ""
        project: ""
        credentialsJson: ""          # Service account JSON
```

## Helm Release

```yaml
tap:
  release:
    repo: https://helm.kubeshark.com  # Helm chart repository
    name: kubeshark                    # Release name
    namespace: default                 # Release namespace
    helmChartPath: ""                  # Path to local chart (overrides repo)
```

## Storage

```yaml
tap:
  persistentStorage: false           # Enable PVC for worker data
  persistentStorageStatic: false     # Static provisioning
  persistentStoragePvcVolumeMode: FileSystem  # FileSystem or Block
  efsFileSytemIdAndPath: ""          # EFS file system ID (EKS)
  secrets: []                        # Additional secrets to mount
  storageLimit: 10Gi                 # Max storage per node
  storageClass: standard             # Default storage class
```

## Resources

```yaml
tap:
  resources:
    hub:
      limits:
        cpu: "0"                     # 0 = no limit
        memory: 5Gi
      requests:
        cpu: 50m
        memory: 50Mi
    sniffer:
      limits:
        cpu: "0"
        memory: 5Gi
      requests:
        cpu: 50m
        memory: 50Mi
    tracer:
      limits:
        cpu: "0"
        memory: 5Gi
      requests:
        cpu: 50m
        memory: 50Mi
```

## Health Probes

```yaml
tap:
  probes:
    hub:
      initialDelaySeconds: 5
      periodSeconds: 5
      successThreshold: 1
      failureThreshold: 3
    sniffer:
      initialDelaySeconds: 5
      periodSeconds: 5
      successThreshold: 1
      failureThreshold: 3
```

## TLS & Service Mesh

```yaml
tap:
  serviceMesh: true                  # Capture mTLS traffic (service mesh)
  tls: true                          # Capture OpenSSL/Go TLS traffic
  disableTlsLog: true                # Suppress TLS debug logging
  packetCapture: best                # Capture method: best, af_packet, pcap
```

## Labels, Annotations & Scheduling

```yaml
tap:
  labels: {}                         # Additional labels for all pods
  annotations: {}                    # Additional annotations for all pods
  nodeSelectorTerms:
    hub:                             # Hub pod node selector
      - matchExpressions:
        - key: kubernetes.io/os
          operator: In
          values: [linux]
    workers:                         # Worker DaemonSet node selector
      - matchExpressions:
        - key: kubernetes.io/os
          operator: In
          values: [linux]
    front:                           # Frontend pod node selector
      - matchExpressions:
        - key: kubernetes.io/os
          operator: In
          values: [linux]
  tolerations:
    hub: []
    workers:
      - operator: Exists
        effect: NoExecute            # Workers tolerate NoExecute by default
    front: []
  priorityClass: ""                  # PriorityClassName for pods
```

## Authentication

```yaml
tap:
  auth:
    enabled: false
    type: saml                       # Only SAML supported currently
    roles:
      admin:
        filter: ""                   # KFL filter restricting visible traffic
        canDownloadPCAP: true
        canUseScripting: true
        scriptingPermissions:
          canSave: true
          canActivate: true
          canDelete: true
        canUpdateTargetedPods: true
        canStopTrafficCapturing: true
        canControlDissection: true
        showAdminConsoleLink: true
    rolesClaim: role                 # SAML attribute for role mapping
    defaultRole: ""                  # Role for users without a role claim
    defaultFilter: ""                # Default KFL filter for all users
    saml:
      idpMetadataUrl: ""             # SAML IdP metadata URL
      x509crt: ""                    # SP certificate
      x509key: ""                    # SP private key
```

## Ingress

```yaml
tap:
  ingress:
    enabled: false
    className: ""                    # nginx, alb, traefik, etc.
    host: ks.svc.cluster.local
    path: /
    tls: []                          # TLS configuration
    annotations: {}                  # Ingress annotations
```

## Protocol Dissectors

```yaml
tap:
  enabledDissectors:
    - amqp
    - dns
    - http
    - icmp
    - kafka
    - mongodb
    - mysql
    - postgresql
    - redis
    - ws
    - ldap
    - radius
    - diameter
    - udp-flow
    - tcp-flow
    - udp-conn
    - tcp-conn
  portMapping:                       # Default port-to-protocol mappings
    http: [80, 443, 8080]
    amqp: [5671, 5672]
    kafka: [9092]
    mongodb: [27017]
    mysql: [3306]
    postgresql: [5432]
    redis: [6379]
    ldap: [389]
    diameter: [3868]
  customMacros:
    https: "tls and (http or http2)"
```

## Networking & Security

```yaml
tap:
  hostNetwork: true                  # Use host network (required for capture)
  ipv6: true                         # Enable IPv6 support
  mountBpf: true                     # Mount BPF filesystem
  securityContext:
    privileged: true
    appArmorProfile:
      type: ""
      localhostProfile: ""
    seLinuxOptions:
      level: ""
      role: ""
      type: ""
      user: ""
    capabilities:
      networkCapture: [NET_RAW, NET_ADMIN]
      serviceMeshCapture: [SYS_ADMIN, SYS_PTRACE, DAC_OVERRIDE]
      ebpfCapture: [SYS_ADMIN, SYS_PTRACE, SYS_RESOURCE, IPC_LOCK]
```

## Dashboard

```yaml
tap:
  dashboard:
    streamingType: connect-rpc
    completeStreamingEnabled: true
    clusterWideMapEnabled: false
    entriesLimit: "300000"
  routing:
    front:
      basePath: ""                   # Base path for reverse proxy
```

## Scripting

```yaml
scripting:
  enabled: false
  env: {}                            # Environment variables for scripts
  source: ""                         # Git repo for scripts
  sources: []                        # Multiple script sources
  watchScripts: true                 # Watch for script changes
  active: []                         # Active scripts
  console: true                      # Enable script console
```

## Misc

```yaml
tap:
  dryRun: false                      # Preview targeted pods without deploying
  debug: false                       # Enable debug mode
  telemetry:
    enabled: true                    # Anonymous usage telemetry
  resourceGuard:
    enabled: false                   # Resource usage guard
  watchdog:
    enabled: false                   # Watchdog process
  gitops:
    enabled: false                   # GitOps mode
  defaultFilter: ""                  # Default KFL display filter
  globalFilter: ""                   # Global KFL filter (cannot be overridden)
  dns:
    nameservers: []                  # Custom DNS nameservers
    searches: []                     # Custom DNS search domains
    options: []                      # Custom DNS options
  misc:
    jsonTTL: 5m                      # TTL for JSON entries
    pcapTTL: "0"                     # TTL for PCAP files (0 = no TTL)
    trafficSampleRate: 100           # Traffic sampling rate (1-100)
    resolutionStrategy: auto         # IP resolution: auto, dns, k8s
    detectDuplicates: false          # Detect duplicate packets
    staleTimeoutSeconds: 30          # Timeout for stale connections
    tcpFlowTimeout: 1200             # TCP flow idle timeout (seconds)
    udpFlowTimeout: 1200             # UDP flow idle timeout (seconds)

headless: false                      # Suppress browser auto-open
license: ""                          # Kubeshark Pro license key
timezone: ""                         # Override timezone
logLevel: warning                    # Log level: debug, info, warning, error

kube:
  configPath: ""                     # Custom kubeconfig path
  context: ""                        # Kubernetes context name
```
