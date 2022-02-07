#!/bin/bash

PREFIX=$HOME/local/bin
VERSION=v1.22.0

echo "Attempting to install minikube and assorted tools to $PREFIX"

if ! [ -x "$(command -v kubectl)" ]; then
  echo "Installing kubectl version $VERSION"
  curl -LO "https://storage.googleapis.com/kubernetes-release/release/$VERSION/bin/linux/amd64/kubectl"
  chmod +x kubectl
  mv kubectl "$PREFIX"
else
  echo "kubetcl is already installed"
fi

if ! [ -x "$(command -v minikube)" ]; then
  echo "Installing minikube version $VERSION"
  curl -Lo minikube https://storage.googleapis.com/minikube/releases/$VERSION/minikube-linux-amd64
  chmod +x minikube
  mv minikube "$PREFIX"
else
  echo "minikube is already installed"
fi

echo "Starting minikube..."
minikube start

echo "Creating mizu tests namespaces"
kubectl create namespace mizu-tests
kubectl create namespace mizu-tests2

echo "Creating httpbin deployments"
kubectl create deployment httpbin --image=kennethreitz/httpbin -n mizu-tests
kubectl create deployment httpbin2 --image=kennethreitz/httpbin -n mizu-tests

kubectl create deployment httpbin --image=kennethreitz/httpbin -n mizu-tests2

echo "Creating redis deployment"
kubectl create deployment redis --image=redis -n mizu-tests

echo "Creating httpbin services"
kubectl expose deployment httpbin --type=NodePort --port=80 -n mizu-tests
kubectl expose deployment httpbin2 --type=NodePort --port=80 -n mizu-tests

kubectl expose deployment httpbin --type=NodePort --port=80 -n mizu-tests2

echo "Creating redis service"
kubectl expose deployment redis --type=LoadBalancer --port=6379 -n mizu-tests

echo "Starting proxy"
kubectl proxy --port=8080 &

echo "Starting tunnel"
minikube tunnel &

echo "Setting minikube docker env"
eval $(minikube docker-env)

echo "Build agent image"
make build-docker-ci

echo "Build cli"
make build-cli-ci

kubectl get services -n mizu-tests
