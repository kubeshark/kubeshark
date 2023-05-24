#!/bin/bash

__dir="$(cd -P -- "$(dirname -- "$0")" && pwd -P)"

helm repo add jetstack https://charts.jetstack.io
helm repo update
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.9.1/cert-manager.crds.yaml
helm install \
cert-manager jetstack/cert-manager \
--namespace cert-manager \
--create-namespace \
--version v1.9.1

kubectl apply -f ${__dir}/cluster-issuer.yaml
kubectl apply -f ${__dir}/certificate.yaml
