# Helm Chart of Kubeshark

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

See the generated templates

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
