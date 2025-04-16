# Helm Chart of Kubeshark

## Official

Add the Helm repo for Kubeshark:

```shell
helm repo add kubeshark https://helm.kubeshark.co
```

then install Kubeshark:

```shell
helm install kubeshark kubeshark/kubeshark
```

## Local

Clone the repo:

```shell
git clone git@github.com:kubeshark/kubeshark.git --depth 1
cd kubeshark/helm-chart
```

In case you want to clone a specific tag of the repo (e.g. `v52.3.59`):

```shell
git clone git@github.com:kubeshark/kubeshark.git --depth 1 --branch <tag>
cd kubeshark/helm-chart
```
> See the list of available tags here: https://github.com/kubeshark/kubeshark/tags

Render the templates

```shell
helm template .
```

Install Kubeshark:

```shell
helm install kubeshark .
```

Uninstall Kubeshark:

```shell
helm uninstall kubeshark
```

## Port-forward

Do the port forwarding:

```shell
kubectl port-forward service/kubeshark-front 8899:80
```

Visit [localhost:8899](http://localhost:8899)

You can also use `kubeshark proxy` for a more stable port-forward connection.

## Add a License Key

When it's necessary, you can use:

```shell
--set license=YOUR_LICENSE_GOES_HERE
```

Get your license from Kubeshark's [Admin Console](https://console.kubeshark.co/).

## Installing with Ingress (EKS) enabled

```shell
helm install kubeshark kubeshark/kubeshark -f values.yaml
```

Set this `value.yaml`:
```shell
tap:
  ingress:
    enabled: true
    className: "alb"
    host: ks.example.com
    tls: []
    annotations:
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-east-1:7..8:certificate/b...65c
      alb.ingress.kubernetes.io/target-type: ip
      alb.ingress.kubernetes.io/scheme: internet-facing
```

## Disabling IPV6

Not all have IPV6 enabled, hence this has to be disabled as follows:

```shell
helm install kubeshark kubeshark/kubeshark \
  --set tap.ipv6=false
```

## Prometheus Metrics

Please refer to [metrics](./metrics.md) documentation for details.

## Override Tag, Tags, Images

In addition to using a private registry, you can further override the images' tag, specific image tags and specific image names.

Example for overriding image names:

```yaml
  docker:
    overrideImage: 
      worker: docker.io/kubeshark/worker:v52.3.87
      front:  docker.io/kubeshark/front:v52.3.87
      hub:    docker.io/kubeshark/hub:v52.3.87
```

## Configuration

| Parameter                                 | Description                                   | Default                                                 |
|-------------------------------------------|-----------------------------------------------|---------------------------------------------------------|
| `tap.docker.registry`                     | Docker registry to pull from                  | `docker.io/kubeshark`                                   |
| `tap.docker.tag`                          | Tag of the Docker images                      | `latest`                                                |
| `tap.docker.tagLocked`                    | Lock the Docker image tags to prevent automatic upgrades to the latest branch image version. | `true`   |
| `tap.docker.tagLocked`                    | If `false` - use latest minor tag             | `true`                                                  |
| `tap.docker.imagePullPolicy`              | Kubernetes image pull policy                  | `Always`                                                |
| `tap.docker.imagePullSecrets`             | Kubernetes secrets to pull the images         | `[]`                                                    |
| `tap.docker.overrideImage`                | Can be used to directly override image names  | `""`                                                    |
| `tap.docker.overrideTag`                  | Can be used to override image tags            | `""`                                                    |
| `tap.proxy.hub.srvPort`                   | Hub server port. Change if already occupied.  | `8898`                                                  |
| `tap.proxy.worker.srvPort`                | Worker server port. Change if already occupied.| `48999`                                                |
| `tap.proxy.front.port`                    | Front service port. Change if already occupied.| `8899`                                                 |
| `tap.proxy.host`                          | Change to 0.0.0.0 top open up to the world.   | `127.0.0.1`                                             |
| `tap.regex`                               | Target (process traffic from) pods that match regex | `.*`                                              |
| `tap.namespaces`                          | Target pods in namespaces                     | `[]`                                                    |
| `tap.excludedNamespaces`                  | Exclude pods in namespaces                    | `[]`                                                    |
| `tap.bpfOverride`                         | When using AF_PACKET as a traffic capture backend, override any existing pod targeting rules and set explicit BPF expression (e.g. `net 0.0.0.0/0`).                                                          | `[]`                                                    |
| `tap.stopped`                             | Set to `false` to have traffic processing start automatically. When set to `true`, traffic processing is stopped by default, resulting in almost no resource consumption (e.g. Kubeshark is dormant). This property can be dynamically control via the dashboard.      | `false`                                                                                                                                                |
| `tap.release.repo`                        | URL of the Helm chart repository              | `https://helm.kubeshark.co`                             |
| `tap.release.name`                        | Helm release name                             | `kubeshark`                                             |
| `tap.release.namespace`                   | Helm release namespace                        | `default`                                               |
| `tap.persistentStorage`                   | Use `persistentVolumeClaim` instead of `emptyDir` | `false`                                             |
| `tap.persistentStorageStatic`             | Use static persistent volume provisioning (explicitly defined `PersistentVolume` ) | `false`            |
| `tap.persistentStoragePvcVolumeMode` | Set the pvc volume mode (Filesystem\|Block) | `Filesystem` |
| `tap.efsFileSytemIdAndPath`               | [EFS file system ID and, optionally, subpath and/or access point](https://github.com/kubernetes-sigs/aws-efs-csi-driver/blob/master/examples/kubernetes/access_points/README.md) `<FileSystemId>:<Path>:<AccessPointId>`  | ""                             |
| `tap.storageLimit`                        | Limit of either the `emptyDir` or `persistentVolumeClaim` | `500Mi`                                     |
| `tap.storageClass`                        | Storage class of the `PersistentVolumeClaim`          | `standard`                                      |
| `tap.dryRun`                              | Preview of all pods matching the regex, without tapping them    | `false`                               |
| `tap.dnsConfig.nameservers`               | Nameservers to use for DNS resolution          | `[]`                                                    |
| `tap.dnsConfig.searches`                  | Search domains to use for DNS resolution       | `[]`                                                    |
| `tap.dnsConfig.options`                   | DNS options to use for DNS resolution          | `[]`                                                    |
| `tap.resources.hub.limits.cpu`            | CPU limit for hub                             | `""`  (no limit)                                                 |
| `tap.resources.hub.limits.memory`         | Memory limit for hub                          | `5Gi`                                                |
| `tap.resources.hub.requests.cpu`          | CPU request for hub                           | `50m`                                                   |
| `tap.resources.hub.requests.memory`       | Memory request for hub                        | `50Mi`                                                  |
| `tap.resources.sniffer.limits.cpu`        | CPU limit for sniffer                         | `""`  (no limit)                                                    |
| `tap.resources.sniffer.limits.memory`     | Memory limit for sniffer                      | `3Gi`                                                |
| `tap.resources.sniffer.requests.cpu`      | CPU request for sniffer                       | `50m`                                                   |
| `tap.resources.sniffer.requests.memory`   | Memory request for sniffer                    | `50Mi`                                                  |
| `tap.resources.tracer.limits.cpu`         | CPU limit for tracer                          | `""`  (no limit)                                                     |
| `tap.resources.tracer.limits.memory`      | Memory limit for tracer                       | `3Gi`                                                |
| `tap.resources.tracer.requests.cpu`       | CPU request for tracer                        | `50m`                                                   |
| `tap.resources.tracer.requests.memory`    | Memory request for tracer                     | `50Mi`                                                  |
| `tap.probes.hub.initialDelaySeconds`      | Initial delay before probing the hub         | `15`                                                    |
| `tap.probes.hub.periodSeconds`            | Period between probes for the hub             | `10`                                                    |
| `tap.probes.hub.successThreshold`         | Number of successful probes before considering the hub healthy | `1`                                        |
| `tap.probes.hub.failureThreshold`         | Number of failed probes before considering the hub unhealthy | `3`                                           |
| `tap.probes.sniffer.initialDelaySeconds`  | Initial delay before probing the sniffer     | `15`                                                    |
| `tap.probes.sniffer.periodSeconds`        | Period between probes for the sniffer         | `10`                                                    |
| `tap.probes.sniffer.successThreshold`     | Number of successful probes before considering the sniffer healthy | `1`                                    |
| `tap.probes.sniffer.failureThreshold`     | Number of failed probes before considering the sniffer unhealthy | `3`                                       |
| `tap.serviceMesh`                         | Capture traffic from service meshes like Istio, Linkerd, Consul, etc.          | `true`                                                  |
| `tap.tls`                                 | Capture the encrypted/TLS traffic from cryptography libraries like OpenSSL                         | `true`                                                  |
| `tap.disableTlsLog`                       | Suppress logging for TLS/eBPF                 | `true`                                                 |
| `tap.labels`                              | Kubernetes labels to apply to all Kubeshark resources  | `{}`                                                    |
| `tap.annotations`                         | Kubernetes annotations to apply to all Kubeshark resources | `{}`                                                |
| `tap.nodeSelectorTerms.workers`                   | Node selector terms for workers components                       | `[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]}]}]` |
| `tap.nodeSelectorTerms.hub`                   | Node selector terms for hub component                 | `[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]}]}]` |
| `tap.nodeSelectorTerms.front`                   | Node selector terms for front-end component                         | `[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]}]}]` |
| `tap.tolerations.workers`                  | Tolerations for workers components                         | `[ {"operator": "Exists", "effect": "NoExecute"}` |
| `tap.tolerations.hub`                  | Tolerations for hub component                         | `[]` |
| `tap.tolerations.front`                  | Tolerations for front-end component                         | `[]` |
| `tap.auth.enabled`                        | Enable authentication                         | `false`                                                 |
| `tap.auth.type`                           | Authentication type (1 option available: `saml`)      | `saml`                                              |
| `tap.auth.approvedEmails`                 | List of approved email addresses for authentication              | `[]`                                                    |
| `tap.auth.approvedDomains`                | List of approved email domains for authentication                | `[]`                                                    |
| `tap.auth.saml.idpMetadataUrl`                    | SAML IDP metadata URL <br/>(effective, if `tap.auth.type = saml`)                                  | ``                                                      |
| `tap.auth.saml.x509crt`                   | A self-signed X.509 `.cert` contents <br/>(effective, if `tap.auth.type = saml`)          | ``                                                      |
| `tap.auth.saml.x509key`                   | A self-signed X.509 `.key` contents <br/>(effective, if `tap.auth.type = saml`)           | ``                                                      |
| `tap.auth.saml.roleAttribute`             | A SAML attribute name corresponding to user's authorization role <br/>(effective, if `tap.auth.type = saml`)  | `role` |
| `tap.auth.saml.roles`                     | A list of SAML authorization roles and their permissions <br/>(effective, if `tap.auth.type = saml`)  | `{"admin":{"canDownloadPCAP":true,"canUpdateTargetedPods":true,"canUseScripting":true, "scriptingPermissions":{"canSave":true, "canActivate":true, "canDelete":true}, "canStopTrafficCapturing":true, "filter":"","showAdminConsoleLink":true}}` |
| `tap.ingress.enabled`                     | Enable `Ingress`                                | `false`                                                 |
| `tap.ingress.className`                   | Ingress class name                            | `""`                                                    |
| `tap.ingress.host`                        | Host of the `Ingress`                          | `ks.svc.cluster.local`                                  |
| `tap.ingress.tls`                         | `Ingress` TLS configuration                     | `[]`                                                    |
| `tap.ingress.annotations`                 | `Ingress` annotations                           | `{}`                                                    |
| `tap.routing.front.basePath`             | Set this value to serve `front` under specific base path. Example: `/custompath` (forward slash must be present)         | `""`       |
| `tap.ipv6`                                | Enable IPv6 support for the front-end                        | `true`                                                  |
| `tap.debug`                               | Enable debug mode                             | `false`                                                 |
| `tap.telemetry.enabled`                   | Enable anonymous usage statistics collection           | `true`                                                  |
| `tap.resourceGuard.enabled`               | Enable resource guard worker process, which watches RAM/disk usage and enables/disables traffic capture based on available resources | `false` |
| `tap.sentry.enabled`                      | Enable sending of error logs to Sentry          | `true` (only for qualified users)                                                  |
| `tap.sentry.environment`                      | Sentry environment to label error logs with      | `production`                                                  |
| `tap.defaultFilter`                       | Sets the default dashboard KFL filter (e.g. `http`). By default, this value is set to filter out noisy protocols such as DNS, UDP, ICMP and TCP. The user can easily change this, **temporarily**, in the Dashboard. For a permanent change, you should change this value in the `values.yaml` or `config.yaml` file.        | `"!dns and !error"`                                    |
| `tap.liveConfigMapChangesDisabled`        | If set to `true`, all user functionality (scripting, targeting settings, global & default KFL modification, traffic recording, traffic capturing on/off, protocol dissectors) involving dynamic ConfigMap changes from UI will be disabled     | `false`      |
| `tap.globalFilter`                        | Prepends to any KFL filter and can be used to limit what is visible in the dashboard. For example, `redact("request.headers.Authorization")` will redact the appropriate field. Another example `!dns` will not show any DNS traffic.      | `""`                                        |
| `tap.metrics.port`                  | Pod port used to expose Prometheus metrics          | `49100`                                                  |
| `tap.enabledDissectors`                   | This is an array of strings representing the list of supported protocols. Remove or comment out redundant protocols (e.g., dns).| The default list excludes: `udp` and `tcp` |
| `tap.mountBpf`                            | BPF filesystem needs to be mounted for eBPF to work properly. This helm value determines whether Kubeshark will attempt to mount the filesystem. This option is not required if filesystem is already mounts. │ `true`|
| `tap.gitops.enabled`                          | Enable GitOps functionality. This will allow you to use GitOps to manage your Kubeshark configuration. | `false` |
| `logs.file`                               | Logs dump path                      | `""`                                                    |
| `pcapdump.enabled`                        | Enable recording of all traffic captured according to other parameters. Whatever Kubeshark captures, considering pod targeting rules, will be stored in pcap files ready to be viewed by tools                 | `true`                                                                                                  |
| `pcapdump.maxTime`                        | The time window into the past that will be stored. Older traffic will be discarded.  | `2h`  |
| `pcapdump.maxSize`                        | The maximum storage size the PCAP files will consume. Old files that cause to surpass storage consumption will get discarded.   | `500MB`  |
| `kube.configPath`                         | Path to the `kubeconfig` file (`$HOME/.kube/config`)            | `""`                                                    |
| `kube.context`                            | Kubernetes context to use for the deployment  | `""`                                                    |
| `dumpLogs`                                | Enable dumping of logs         | `false`                                                 |
| `headless`                                | Enable running in headless mode               | `false`                                                 |
| `license`                                 | License key for the Pro/Enterprise edition    | `""`                                                    |
| `scripting.env`                           | Environment variables for the scripting      | `{}`                                                    |
| `scripting.source`                        | Source directory of the scripts                | `""`                                                    |
| `scripting.watchScripts`                  | Enable watch mode for the scripts in source directory          | `true`                                                  |
| `timezone`                                | IANA time zone applied to time shown in the front-end | `""` (local time zone applies) |
| `supportChatEnabled`                      | Enable real-time support chat channel based on Intercom | `false` |
| `internetConnectivity`                    | Turns off API requests that are dependant on Internet connectivity such as `telemetry` and `online-support`. | `true` |

KernelMapping pairs kernel versions with a
                            DriverContainer image. Kernel versions can be matched
                            literally or using a regular expression

# Installing with SAML enabled

### Prerequisites:

##### 1. Generate X.509 certificate & key (TL;DR: https://ubuntu.com/server/docs/security-certificates)

**Example:**
```
openssl genrsa -out mykey.key 2048
openssl req -new -key mykey.key -out mycsr.csr
openssl x509 -signkey mykey.key -in mycsr.csr -req -days 365 -out mycert.crt
```

**What you get:**
- `mycert.crt` - use it for `tap.auth.saml.x509crt`
- `mykey.key` - use it for `tap.auth.saml.x509crt`

##### 2. Prepare your SAML IDP

You should set up the required SAML IDP (Google, Auth0, your custom IDP, etc.)

During setup, an IDP provider will typically request to enter:
- Metadata URL
- ACS URL (Assertion Consumer Service URL, aka Callback URL)
- SLO URL (Single Logout URL)

Correspondingly, you will enter these (if you run the most default Kubeshark setup):
- [http://localhost:8899/saml/metadata](http://localhost:8899/saml/metadata)
- [http://localhost:8899/saml/acs](http://localhost:8899/saml/acs)
- [http://localhost:8899/saml/slo](http://localhost:8899/saml/slo)

Otherwise, if you have `tap.ingress.enabled == true`, change protocol & domain respectively - showing example domain:
- [https://kubeshark.example.com/saml/metadata](https://kubeshark.example.com/saml/metadata)
- [https://kubeshark.example.com/saml/acs](https://kubeshark.example.com/saml/acs)
- [https://kubeshark.example.com/saml/slo](https://kubeshark.example.com/saml/slo)

```shell
helm install kubeshark kubeshark/kubeshark -f values.yaml
```

Set this `value.yaml`:
```shell
tap:
  auth:
    enabled: true
    type: saml
    saml:
      idpMetadataUrl: "https://ti..th0.com/samlp/metadata/MpWiDCM..qdnDG"
      x509crt: |
        -----BEGIN CERTIFICATE-----
        MIIDlTCCAn0CFFRUzMh+dZvp+FvWd4gRaiBVN8EvMA0GCSqGSIb3DQEBCwUAMIGG
        MSQwIgYJKoZIhvcNAQkBFhV3ZWJtYXN0ZXJAZXhhbXBsZS5jb20wHhcNMjMxMjI4
        ........<redacted: please, generate your own X.509 cert>........
        ZMzM7YscqZwoVhTOhrD4/5nIfOD/hTWG/MBe2Um1V1IYF8aVEllotTKTgsF6ZblA
        miCOgl6lIlZy
        -----END CERTIFICATE-----
      x509key: |
        -----BEGIN PRIVATE KEY-----
        MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDlgDFKsRHj+mok
        euOF0IpwToOEpQGtafB75ytv3psD/tQAzEIug+rkDriVvsfcvafj0qcaTeYvnCoz
        ........<redacted: please, generate your own X.509 key>.........
        sUpBCu0E3nRJM/QB2ui5KhNR7uvPSL+kSsaEq19/mXqsL+mRi9aqy2wMEvUSU/kt
        UaV5sbRtTzYLxpOSQyi8CEFA+A==
        -----END PRIVATE KEY-----
```

# Installing with Dex OIDC authentication

[**Click here to see full docs**](https://docs.kubeshark.co/en/saml#installing-with-oidc-enabled-dex-idp).

Choose this option, if **you already have a running instance** of Dex in your cluster & 
you want to set up Dex OIDC authentication for Kubeshark users.

Kubeshark supports authentication using [Dex - A Federated OpenID Connect Provider](https://dexidp.io/).
Dex is an abstraction layer designed for integrating a wide variety of Identity Providers.

**Requirement:**
Your Dex IdP must have a publicly accessible URL.

### Pre-requisites:

**1. If you configured Ingress for Kubeshark:**

(see section: "Installing with Ingress (EKS) enabled")

OAuth2 callback URL is: <br/>
`https://<kubeshark-ingress-hostname>/api/oauth2/callback`

**2. If you did not configure Ingress for Kubeshark:**

OAuth2 callback URL is: <br/>
`http://0.0.0.0:8899/api/oauth2/callback`

Use chosen OAuth2 callback URL to replace `<your-kubeshark-host>` in Step 3.

**3. Add this static client to your Dex IdP configuration (`config.yaml`):**
```yaml
staticClients:
   - id: kubeshark
     secret: create your own client password
     name: Kubeshark
     redirectURIs:
     - https://<your-kubeshark-host>/api/oauth2/callback
```

**Final step:**

Add these helm values to set up OIDC authentication powered by your Dex IdP:

```yaml
# values.yaml

tap: 
  auth:
    enabled: true
    type: dex
    dexOidc:
      issuer: <put Dex IdP issuer URL here>
      clientId: kubeshark
      clientSecret: create your own client password
      refreshTokenLifetime: "3960h" # 165 days
      oauth2StateParamExpiry: "10m"
      bypassSslCaCheck: false
```

---

**Note:**<br/>
Set `tap.auth.dexOidc.bypassSslCaCheck: true`
to allow Kubeshark communication with Dex IdP having an unknown SSL Certificate Authority.

This setting allows you to prevent such SSL CA-related errors:<br/>
`tls: failed to verify certificate: x509: certificate signed by unknown authority`

---

Once you run `helm install kubeshark kubeshark/kubeshark -f ./values.yaml`, Kubeshark will be installed with (Dex) OIDC authentication enabled.

---

# Installing your own Dex IdP along with Kubeshark

Choose this option, if **you need to deploy an instance of Dex IdP** along with Kubeshark & 
set up Dex OIDC authentication for Kubeshark users.

Depending on Ingress enabled/disabled, your Dex configuration might differ.

**Requirement:**
Please, configure Ingress using `tap.ingress` for your Kubeshark installation. For example:

```yaml
tap:
  ingress:
    enabled: true
    className: "alb"
    host: ks.example.com
    tls: []
    annotations:
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-east-1:7..8:certificate/b...65c
      alb.ingress.kubernetes.io/target-type: ip
      alb.ingress.kubernetes.io/scheme: internet-facing
```

The following Dex settings will have these values:

| Setting                                               | Value                                        |
|-------------------------------------------------------|----------------------------------------------|
| `tap.auth.dexOidc.issuer`                             | `https://ks.example.com/dex`                 |
| `tap.auth.dexConfig.issuer`                           | `https://ks.example.com/dex`                 |
| `tap.auth.dexConfig.staticClients -> redirectURIs`    | `https://ks.example.com/api/oauth2/callback` |
| `tap.auth.dexConfig.connectors -> config.redirectURI` | `https://ks.example.com/dex/callback`        |

---

### Before proceeding with Dex IdP installation:

Please, make sure to prepare the following things first.

1. Choose **[Connectors](https://dexidp.io/docs/connectors/)** to enable in Dex IdP. 
   - i.e. how many kind of "Log in with ..." options you'd like to offer your users
   - You will need to specify connectors in `tap.auth.dexConfig.connectors`
2. Choose type of **[Storage](https://dexidp.io/docs/configuration/storage/)** to use in Dex IdP. 
   - You will need to specify storage settings in `tap.auth.dexConfig.storage`
   - default: `memory`
3. Decide on the OAuth2 `?state=` param expiration time:
   - field: `tap.auth.dexOidc.oauth2StateParamExpiry`
   - default: `10m` (10 minutes)
   - valid time units are `s`, `m`, `h`
4. Decide on the refresh token expiration:
    - field 1: `tap.auth.dexOidc.expiry.refreshTokenLifetime`
    - field 2: `tap.auth.dexConfig.expiry.refreshTokens.absoluteLifetime`
    - default: `3960h` (165 days)
    - valid time units are `s`, `m`, `h`
5. Create a unique & secure password to set in these fields:
    - field 1: `tap.auth.dexOidc.clientSecret`
    - field 2: `tap.auth.dexConfig.staticClients -> secret`
    - password must be the same for these 2 fields
6. Discover more possibilities of **[Dex Configuration](https://dexidp.io/docs/configuration/)**
   - if you decide to include more configuration options, make sure to add them into `tap.auth.dexConfig`
---

### Once you are ready with all the points described above:

Use these helm `values.yaml` fields to:
- Deploy your own instance of Dex IdP along with Kubeshark
- Enable OIDC authentication for Kubeshark users

Make sure to:
- Replace `<your-ingress-hostname>` with a correct Kubeshark Ingress host (`tap.auth.ingress.host`).
  - refer to section **Installing with Ingress (EKS) enabled** to find out how you can configure Ingress host.

Helm `values.yaml`:
```yaml
tap: 
  auth:
    enabled: true
    type: dex
    dexOidc:
      issuer: https://<your-ingress-hostname>/dex
      
      # Client ID/secret must be taken from `tap.auth.dexConfig.staticClients -> id/secret`
      clientId: kubeshark
      clientSecret: create your own client password
      
      refreshTokenLifetime: "3960h" # 165 days
      oauth2StateParamExpiry: "10m"
      bypassSslCaCheck: false
    dexConfig:
      # This field is REQUIRED!
      # 
      # The base path of Dex and the external name of the OpenID Connect service.
      # This is the canonical URL that all clients MUST use to refer to Dex. If a
      # path is provided, Dex's HTTP service will listen at a non-root URL.
      issuer: https://<your-ingress-hostname>/dex
        
      # Expiration configuration for tokens, signing keys, etc.
      expiry:
        refreshTokens:
          validIfNotUsedFor: "2160h" # 90 days
          absoluteLifetime: "3960h"  # 165 days

      # This field is REQUIRED!
      # 
      # The storage configuration determines where Dex stores its state.
      # See the documentation (https://dexidp.io/docs/storage/) for further information.
      storage:
        type: memory

      # This field is REQUIRED!
      # 
      # Attention: 
      # Do not change this field and its values.
      # This field is required for internal Kubeshark-to-Dex communication.
      #
      # HTTP service configuration
      web:
        http: 0.0.0.0:5556

      # This field is REQUIRED!
      #
      # Attention: 
      # Do not change this field and its values.
      # This field is required for internal Kubeshark-to-Dex communication.
      #
      # Telemetry configuration
      telemetry:
        http: 0.0.0.0:5558

      # This field is REQUIRED!
      #
      # Static clients registered in Dex by default.
      staticClients:
        - id: kubeshark
          secret: create your own client password
          name: Kubeshark
          redirectURIs:
          - https://<your-ingress-hostname>/api/oauth2/callback

      # Enable the password database.
      # It's a "virtual" connector (identity provider) that stores
      # login credentials in Dex's store.
      enablePasswordDB: true

      # Connectors are used to authenticate users against upstream identity providers.
      # See the documentation (https://dexidp.io/docs/connectors/) for further information.
      #
      # Attention: 
      # When you define a new connector, `config.redirectURI` must be: 
      # https://<your-ingress-hostname>/dex/callback
      # 
      # Example with Google connector:
      # connectors:
      #  - type: google
      #    id: google
      #    name: Google
      #    config:
      #      clientID: your Google Cloud Auth app client ID
      #      clientSecret: your Google Auth app client ID
      #      redirectURI: https://<your-ingress-hostname>/dex/callback
      connectors: []
```
