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

## Installing with Ingress Enabled

```shell
helm install kubeshark kubeshark/kubeshark \
  --set tap.ingress.enabled=true \
  --set tap.ingress.host=ks.svc.cluster.local \
  --set-json='tap.ingress.approveddomains=["gmail.com"]' \
  --set license=LICENSE_GOES_HERE \
  --set-json 'tap.annotations={ "eks.amazonaws.com/role-arn" : "arn:aws:iam::7...0:role/s3-role" }'
```

You can get your license [here](https://console.kubeshark.co/).

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
| `tap.proxy.worker.srvport`                | Worker server port                           | `8897`                                                  |
| `tap.proxy.hub.port`                      | Hub service port                              | `8898`                                                  |
| `tap.proxy.hub.srvport`                   | Hub server port                   | `8898`                                                  |
| `tap.proxy.front.port`                    | Front-facing service port                     | `8899`                                                  |
| `tap.proxy.host`                          | Proxy server's IP                                   | `127.0.0.1`                                             |
| `tap.namespaces`                          | List of namespaces for the traffic capture                 | `[]`                                                    |
| `tap.release.repo`                        | URL of the Helm chart repository             | `https://helm.kubeshark.co`                             |
| `tap.release.name`                        | Helm release name                          | `kubeshark`                                             |
| `tap.release.namespace`                   | Helm release namespace                | `default`                                               |
| `tap.persistentstorage`                   | Use `persistentVolumeClaim` instead of `emptyDir` | `false`                                                |
| `tap.storagelimit`                        | Limit of either the `emptyDir` or `persistentVolumeClaim`                  | `500Mi`                                                 |
| `tap.storageclass`                        | Storage class of the `PersistentVolumeClaim`          | `standard`                                              |
| `tap.dryrun`                              | Preview of all pods matching the regex, without tapping them                    | `false`                                                 |
| `tap.pcap`                                |                                               | `""`                                                    |
| `tap.resources.worker.limits.cpu`         | CPU limit for worker                          | `750m`                                                  |
| `tap.resources.worker.limits.memory`      | Memory limit for worker                       | `1Gi`                                                   |
| `tap.resources.worker.requests.cpu`       | CPU request for worker                        | `50m`                                                   |
| `tap.resources.worker.requests.memory`    | Memory request for worker                     | `50Mi`                                                  |
| `tap.resources.hub.limits.cpu`            | CPU limit for hub                             | `750m`                                                  |
| `tap.resources.hub.limits.memory`         | Memory limit for hub                          | `1Gi`                                                   |
| `tap.resources.hub.requests.cpu`          | CPU request for hub                           | `50m`                                                   |
| `tap.resources.hub.requests.memory`       | Memory request for hub                        | `50Mi`                                                  |
| `tap.servicemesh`                         | Capture traffic from service meshes like Istio, Linkerd, Consul, etc.          | `true`                                                  |
| `tap.tls`                                 | Capture the encrypted/TLS traffic from cryptography libraries like OpenSSL                         | `true`                                                  |
| `tap.ignoretainted`                       | Whether to ignore tainted nodes               | `false`                                                 |
| `tap.labels`                              | Kubernetes labels to apply to all Kubeshark resources  | `{}`                                                    |
| `tap.annotations`                         | Kubernetes annotations to apply to all Kubeshark resources | `{}`                                                |
| `tap.nodeselectorterms`                   | Node selector terms                           | `[{"matchExpressions":[{"key":"kubernetes.io/os","operator":"In","values":["linux"]}]}]` |
| `tap.auth.enabled`                        | Enable authentication                         | `false`                                                 |
| `tap.auth.approvedemails`                 | List of approved email addresses for authentication              | `[]`                                                    |
| `tap.auth.approveddomains`                | List of approved email domains for authentication                | `[]`                                                    |
| `tap.ingress.enabled`                     | Enable `Ingress`                                | `false`                                                 |
| `tap.ingress.classname`                   | Ingress class name                            | `""`                                                    |
| `tap.ingress.host`                        | Host of the `Ingress`                          | `ks.svc.cluster.local`                                  |
| `tap.ingress.tls`                         | `Ingress` TLS configuration                     | `[]`                                                    |
| `tap.ingress.annotations`                 | `Ingress` annotations                           | `{}`                                                    |
| `tap.ipv6`                                | Enable IPv6 support for the front-end                        | `true`                                                  |
| `tap.debug`                               | Enable debug mode                             | `false`                                                 |
| `tap.nokernelmodule`                      | Do not install `PF_RING` kernel module       | `false`                                                 |
| `tap.telemetry.enabled`                   | Enable anonymous usage statistics collection           | `true`                                                  |
| `logs.file`                               | Logs dump path                      | `""`                                                    |
| `kube.configpath`                         | Path to the `kubeconfig` file (`$HOME/.kube/config`)            | `""`                                                    |
| `kube.context`                            | Kubernetes context to use for the deployment  | `""`                                                    |
| `dumplogs`                                | Enable dumping of logs         | `false`                                                 |
| `headless`                                | Enable running in headless mode               | `false`                                                 |
| `license`                                 | License key for the Pro/Enterprise edition    | `""`                                                    |
| `scripting.env`                           | Environment variables for the scripting      | `{}`                                                    |
| `scripting.source`                        | Source directory of the scripts                | `""`                                                    |
| `scripting.watchscripts`                  | Enable watch mode for the scripts in source directory          | `true`                                                  |
