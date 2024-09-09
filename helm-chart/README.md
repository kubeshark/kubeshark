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


## Increase the Worker's Storage Limit

For example, change from the default 500Mi to 5Gi:

```shell
--set tap.storageLimit=5Gi
```

## Add a License

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

## Metrics

Please refer to [metrics](./metrics.md) documentation for details.

## Configuration

| Parameter                                 | Description                                   | Default                                                 |
|-------------------------------------------|-----------------------------------------------|---------------------------------------------------------|
| `tap.docker.registry`                     | Docker registry to pull from                           | `docker.io/kubeshark`                                   |
| `tap.docker.tag`                          | Tag of the Docker images                              | `latest`                                                |
| `tap.docker.tagLocked`                    | If `false` - use latest minor tag             | `true`                                                  |
| `tap.docker.imagePullPolicy`              | Kubernetes image pull policy                  | `Always`                                                |
| `tap.docker.imagePullSecrets`             | Kubernetes secrets to pull the images      | `[]`                                                    |
| `tap.proxy.worker.srvPort`                | Worker server port                           | `30001`                                                  |
| `tap.proxy.hub.port`                      | Hub service port                              | `8898`                                                  |
| `tap.proxy.hub.srvPort`                   | Hub server port                   | `8898`                                                  |
| `tap.proxy.front.port`                    | Front-facing service port                     | `8899`                                                  |
| `tap.proxy.host`                          | Proxy server's IP                                   | `127.0.0.1`                                             |
| `tap.namespaces`                          | List of namespaces for the traffic capture                 | `[]`                                                    |
| `tap.excludedNamespaces`                  | List of namespaces to explicitly exclude                 | `[]`                                                    |
| `tap.release.repo`                        | URL of the Helm chart repository             | `https://helm.kubeshark.co`                             |
| `tap.release.name`                        | Helm release name                          | `kubeshark`                                             |
| `tap.release.namespace`                   | Helm release namespace                | `default`                                               |
| `tap.persistentStorage`                   | Use `persistentVolumeClaim` instead of `emptyDir` | `false`                                                |
| `tap.persistentStorageStatic`             | Use static persistent volume provisioning (explicitly defined `PersistentVolume` ) | `false`                                                      |
| `tap.efsFileSytemIdAndPath`               | [EFS file system ID and, optionally, subpath and/or access point](https://github.com/kubernetes-sigs/aws-efs-csi-driver/blob/master/examples/kubernetes/access_points/README.md) `<FileSystemId>:<Path>:<AccessPointId>`     | ""                                                           |
| `tap.storageLimit`                        | Limit of either the `emptyDir` or `persistentVolumeClaim`                  | `500Mi`                                                 |
| `tap.storageClass`                        | Storage class of the `PersistentVolumeClaim`          | `standard`                                              |
| `tap.dryRun`                              | Preview of all pods matching the regex, without tapping them                    | `false`                                                 |
| `tap.pcap`                                |                                               | `""`                                                    |
| `tap.resources.worker.limits.cpu`         | CPU limit for worker                          | `750m`                                                  |
| `tap.resources.worker.limits.memory`      | Memory limit for worker                       | `1Gi`                                                   |
| `tap.resources.worker.requests.cpu`       | CPU request for worker                        | `50m`                                                   |
| `tap.resources.worker.requests.memory`    | Memory request for worker                     | `50Mi`                                                  |
| `tap.resources.hub.limits.cpu`            | CPU limit for hub                             | `750m`                                                  |
| `tap.resources.hub.limits.memory`         | Memory limit for hub                          | `1Gi`                                                   |
| `tap.resources.hub.requests.cpu`          | CPU request for hub                           | `50m`                                                   |
| `tap.resources.hub.requests.memory`       | Memory request for hub                        | `50Mi`                                                  |
| `tap.serviceMesh`                         | Capture traffic from service meshes like Istio, Linkerd, Consul, etc.          | `true`                                                  |
| `tap.tls`                                 | Capture the encrypted/TLS traffic from cryptography libraries like OpenSSL                         | `true`                                                  |
| `tap.disableTlsLog`                       | Suppress logging for TLS/eBPF                 | `false`                                                 |
| `tap.ignoreTainted`                       | Whether to ignore tainted nodes               | `false`                                                 |
| `tap.labels`                              | Kubernetes labels to apply to all Kubeshark resources  | `{}`                                                    |
| `tap.annotations`                         | Kubernetes annotations to apply to all Kubeshark resources | `{}`                                                |
| `tap.nodeSelectorTerms`                   | Node selector terms                           | `[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]}]}]` |
| `tap.auth.enabled`                        | Enable authentication                         | `false`                                                 |
| `tap.auth.type`                           | Authentication type (1 option available: `saml`)      | `saml`                                              |
| `tap.auth.approvedEmails`                 | List of approved email addresses for authentication              | `[]`                                                    |
| `tap.auth.approvedDomains`                | List of approved email domains for authentication                | `[]`                                                    |
| `tap.auth.saml.idpMetadataUrl`                    | SAML IDP metadata URL <br/>(effective, if `tap.auth.type = saml`)                                  | ``                                                      |
| `tap.auth.saml.x509crt`                   | A self-signed X.509 `.cert` contents <br/>(effective, if `tap.auth.type = saml`)          | ``                                                      |
| `tap.auth.saml.x509key`                   | A self-signed X.509 `.key` contents <br/>(effective, if `tap.auth.type = saml`)           | ``                                                      |
| `tap.auth.saml.roleAttribute`             | A SAML attribute name corresponding to user's authorization role <br/>(effective, if `tap.auth.type = saml`)  | `role` |
| `tap.auth.saml.roles`                     | A list of SAML authorization roles and their permissions <br/>(effective, if `tap.auth.type = saml`)  | `{"admin":{"canDownloadPCAP":true,"canUpdateTargetedPods":true,"canUseScripting":true, "canStopTrafficCapturing":true, "filter":"","showAdminConsoleLink":true}}` |
| `tap.ingress.enabled`                     | Enable `Ingress`                                | `false`                                                 |
| `tap.ingress.className`                   | Ingress class name                            | `""`                                                    |
| `tap.ingress.host`                        | Host of the `Ingress`                          | `ks.svc.cluster.local`                                  |
| `tap.ingress.tls`                         | `Ingress` TLS configuration                     | `[]`                                                    |
| `tap.ingress.annotations`                 | `Ingress` annotations                           | `{}`                                                    |
| `tap.ipv6`                                | Enable IPv6 support for the front-end                        | `true`                                                  |
| `tap.debug`                               | Enable debug mode                             | `false`                                                 |
| `tap.kernelModule.enabled`                | Use PF_RING kernel module([details](PF_RING.md))      | `false`                                                 |
| `tap.kernelModule.image`                  | Container image containing PF_RING kernel module with supported kernel version([details](PF_RING.md))      | "kubeshark/pf-ring-module:all"                                                 |
| `tap.kernelModule.unloadOnDestroy`        | Create additional container which watches for pod termination and unloads PF_RING kernel module. | `false`|
| `tap.telemetry.enabled`                   | Enable anonymous usage statistics collection           | `true`                                                  |
| `tap.sentry.enabled`                      | Enable sending of error logs to Sentry          | `false`                                                  |
| `tap.sentry.environment`                      | Sentry environment to label error logs with      | `production`                                                  |
| `tap.defaultFilter`                       | Sets the default dashboard KFL filter (e.g. `http`). By default, this value is set to filter out DNS  and TCP entries. The user can easily change this in the Dashboard.         | `"!dns and !tcp"`                                                  |
| `tap.globalFilter`                        | Prepends to any KFL filter and can be used to limit what is visible in the dashboard. For example, `redact("request.headers.Authorization")` will redact the appropriate field. Another example `!dns` will not show any DNS traffic.      | `""`                                        |
| `tap.metrics.port`                  | Pod port used to expose Prometheus metrics          | `49100`                                                  |
| `tap.stopped`                             | A flag indicating whether to start Kubeshark with traffic processing stopped resulting in almost no resource consumption (e.g. Kubeshark is dormant). This property can be dynamically control via the dashboard.         | `true`                                                  |
| `tap.enabledDissectors`                   | This is an array of strings representing the list of supported protocols. Remove or comment out redundant protocols (e.g., dns).| The default list includes: amqp, dns , http, icmp, kafka, redis,sctp, syscall, tcp, ws.  |
| `logs.file`                               | Logs dump path                      | `""`                                                    |
| `kube.configPath`                         | Path to the `kubeconfig` file (`$HOME/.kube/config`)            | `""`                                                    |
| `kube.context`                            | Kubernetes context to use for the deployment  | `""`                                                    |
| `dumpLogs`                                | Enable dumping of logs         | `false`                                                 |
| `headless`                                | Enable running in headless mode               | `false`                                                 |
| `license`                                 | License key for the Pro/Enterprise edition    | `""`                                                    |
| `scripting.env`                           | Environment variables for the scripting      | `{}`                                                    |
| `scripting.source`                        | Source directory of the scripts                | `""`                                                    |
| `scripting.watchScripts`                  | Enable watch mode for the scripts in source directory          | `true`                                                  |
| `timezone`                                | IANA time zone applied to time shown in the front-end | `""` (local time zone applies) |
| `supportChatEnabled`                      | Enable real-time support chat channel based on Intercom | `true` |
| `internetConnectivity`                    | Turns off API requests that are dependant on Internet connectivity such as `telemetry` and `online-support`. | `true` |
| `dissectorsUpdatingEnabled`                     | Turns off UI for enabling/disabling dissectors | `true` |

KernelMapping pairs kernel versions with a
                            DriverContainer image. Kernel versions can be matched
                            literally or using a regular expression

## Installing with SAML enabled

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
      idpMetadataUrl: "https://tiptophelmet.us.auth0.com/samlp/metadata/MpWiDCMMB5ShU1HRnhdb1sHM6VWqdnDG"
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
