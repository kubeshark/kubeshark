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
kubectl port-forward -n kubeshark service/kubeshark-hub 8898:80 & \
kubectl port-forward -n kubeshark service/kubeshark-front 8899:80
```

Visit [localhost:8899](http://localhost:8899)

## Installing with Ingress Enabled

```shell
helm install kubeshark kubeshark/kubeshark \
  --set tap.ingress.enabled=true \
  --set tap.ingress.host=ks.svc.cluster.local \
  --set "tap.ingress.auth.approveddomains={gmail.com}" \
  --set license=LICENSE_GOES_HERE
```

You can get your license [here](https://console.kubeshark.co/).

## Installing with Persistent Storage Enabled

```shell
helm install kubeshark kubeshark/kubeshark \
  --set tap.persistentstorage=true \
  --set license=LICENSE_GOES_HERE
```

You can get your license [here](https://console.kubeshark.co/).

## Disabling IPV6

Not all have IPV6 enabled, hence this has to be disabled as follows:

```shell
helm install kubeshark kubeshark/kubeshark \
  --set tap.ipv6=false
```
