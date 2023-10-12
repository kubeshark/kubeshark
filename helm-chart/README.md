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

## Installing with Ingress (EKS) and Auth Enabled

```shell
helm install kubeshark kubeshark/kubeshark -f values.yaml
```

Set this `value.yaml`:
```shell
tap:
  auth:
    enabled: true
    approvedemails:
    - me@domain.com
    - they@another-domain.com
    approveddomains: []
  ingress:
    enabled: true
    classname: "alb"
    host: demo.mydomain.com
    tls: []
    annotations:
      alb.ingress.kubernetes.io/certificate-arn: arn:aws:acm:us-east-1:7..8:certificate/b...65c
      alb.ingress.kubernetes.io/target-type: ip
      alb.ingress.kubernetes.io/scheme: internet-facing
```

## Add a License

When nessesary, uou can use:

```shell
--set license=YOUR_LICENSE_GOES_HERE
```

Get your license from Kubeshark's [Admin Console](https://console.kubeshark.co/).

## Increase the Worker's Storage Limit

```shell
--set license=YOUR_LICENSE_GOES_HERE
```
 
## Disabling IPV6

Not all have IPV6 enabled, hence this has to be disabled as follows:

```shell
helm install kubeshark kubeshark/kubeshark \
  --set tap.ipv6=false
```
