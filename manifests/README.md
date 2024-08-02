# Manifests

## Apply

Clone the repo:

```shell
git clone git@github.com:kubeshark/kubeshark.git --depth 1
cd kubeshark/manifests
```

To apply the manifests, run:

```shell
kubectl apply -f .
```

To clean up:

```shell
kubectl delete namespace kubeshark
kubectl delete clusterrolebinding kubeshark-cluster-role-binding
kubectl delete clusterrole kubeshark-cluster-role
```

## Accessing

Do the port forwarding:

```shell
kubectl port-forward service/kubeshark-front 8899:80
```

Visit [localhost:8899](http://localhost:8899)
