# Helm Chart of Kubeshark

## Officially

Add the Helm repo for Kubeshark:

```shell
helm repo add kubeshark https://helm.kubeshark.co
```

then install Kubeshark:

```shell
helm install kubeshark kubeshark/kubeshark
```

## Locally

Clone the repo:

```shell
git clone git@github.com:kubeshark/kubeshark.git --depth 1
cd kubeshark/helm-chart
```

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

## Accesing

Do the port forwarding:

```shell
kubectl port-forward service/kubeshark-front 8899:80
```

Visit [localhost:8899](http://localhost:8899)

## Installing with Ingress (EKS) and enable Auth

```shell
helm install kubeshark kubeshark/kubeshark -f values.yaml
```

Set this `value.yaml`:
```shell
tap:
  auth:
    enabled: true
    approvedEmails:
    - john.doe@example.com
    approvedDomains: []
    approvedTenants: []
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

## Add a License

When it's necessary, you can use:

```shell
--set license=YOUR_LICENSE_GOES_HERE
```

Get your license from Kubeshark's [Admin Console](https://console.kubeshark.co/).

## Increase the Worker's Storage Limit

For example, change from the default 500Mi to 1Gi:

```shell
--set tap.storageLimit=1Gi
```
 
## Disabling IPV6

Not all have IPV6 enabled, hence this has to be disabled as follows:

```shell
helm install kubeshark kubeshark/kubeshark \
  --set tap.ipv6=false
```

## Configuration

| Parameter                                 | Description                                   | Default                                                 |
|-------------------------------------------|-----------------------------------------------|---------------------------------------------------------|
| `tap.docker.registry`                     | Docker registry to pull from                           | `docker.io/kubeshark`                                   |
| `tap.docker.tag`                          | Tag of the Docker images                              | `latest`                                                |
| `tap.docker.imagePullPolicy`              | Kubernetes image pull policy                  | `Always`                                                |
| `tap.docker.imagePullSecrets`             | Kubernetes secrets to pull the images      | `[]`                                                    |
| `tap.proxy.worker.srvPort`                | Worker server port                           | `8897`                                                  |
| `tap.proxy.hub.port`                      | Hub service port                              | `8898`                                                  |
| `tap.proxy.hub.srvPort`                   | Hub server port                   | `8898`                                                  |
| `tap.proxy.front.port`                    | Front-facing service port                     | `8899`                                                  |
| `tap.proxy.host`                          | Proxy server's IP                                   | `127.0.0.1`                                             |
| `tap.namespaces`                          | List of namespaces for the traffic capture                 | `[]`                                                    |
| `tap.release.repo`                        | URL of the Helm chart repository             | `https://helm.kubeshark.co`                             |
| `tap.release.name`                        | Helm release name                          | `kubeshark`                                             |
| `tap.release.namespace`                   | Helm release namespace                | `default`                                               |
| `tap.persistentStorage`                   | Use `persistentVolumeClaim` instead of `emptyDir` | `false`                                                |
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
| `tap.ignoreTainted`                       | Whether to ignore tainted nodes               | `false`                                                 |
| `tap.labels`                              | Kubernetes labels to apply to all Kubeshark resources  | `{}`                                                    |
| `tap.annotations`                         | Kubernetes annotations to apply to all Kubeshark resources | `{}`                                                |
| `tap.nodeSelectorTerms`                   | Node selector terms                           | `[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]}]}]` |
| `tap.auth.enabled`                        | Enable authentication                         | `false`                                                 |
| `tap.auth.approvedEmails`                 | List of approved email addresses for authentication              | `[]`                                                    |
| `tap.auth.approvedDomains`                | List of approved email domains for authentication                | `[]`                                                    |
| `tap.ingress.enabled`                     | Enable `Ingress`                                | `false`                                                 |
| `tap.ingress.className`                   | Ingress class name                            | `""`                                                    |
| `tap.ingress.host`                        | Host of the `Ingress`                          | `ks.svc.cluster.local`                                  |
| `tap.ingress.tls`                         | `Ingress` TLS configuration                     | `[]`                                                    |
| `tap.ingress.annotations`                 | `Ingress` annotations                           | `{}`                                                    |
| `tap.ipv6`                                | Enable IPv6 support for the front-end                        | `true`                                                  |
| `tap.debug`                               | Enable debug mode                             | `false`                                                 |
| `tap.noKernelModule`                      | Do not install `PF_RING` kernel module       | `false`                                                 |
| `tap.telemetry.enabled`                   | Enable anonymous usage statistics collection           | `true`                                                  |
| `logs.file`                               | Logs dump path                      | `""`                                                    |
| `kube.configPath`                         | Path to the `kubeconfig` file (`$HOME/.kube/config`)            | `""`                                                    |
| `kube.context`                            | Kubernetes context to use for the deployment  | `""`                                                    |
| `dumpLogs`                                | Enable dumping of logs         | `false`                                                 |
| `headless`                                | Enable running in headless mode               | `false`                                                 |
| `license`                                 | License key for the Pro/Enterprise edition    | `""`                                                    |
| `scripting.env`                           | Environment variables for the scripting      | `{}`                                                    |
| `scripting.source`                        | Source directory of the scripts                | `""`                                                    |
| `scripting.watchScripts`                  | Enable watch mode for the scripts in source directory          | `true`                                                  |
