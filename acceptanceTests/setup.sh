#!/bin/bash
set -e

PREFIX=$HOME/local/bin
VERSION=v1.22.0
TUNNEL_LOG="tunnel.log"
PROXY_LOG="proxy.log"

echo "Attempting to install minikube and assorted tools to $PREFIX"

if ! [ -x "$(command -v kubectl)" ]; then
  echo "Installing kubectl version $VERSION"
  curl -LO "https://storage.googleapis.com/kubernetes-release/release/$VERSION/bin/linux/amd64/kubectl"
  chmod +x kubectl
  mv kubectl "$PREFIX"
else
  echo "kubectl is already installed"
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
kubectl create namespace mizu-tests --dry-run=client -o yaml | kubectl apply -f -
kubectl create namespace mizu-tests2 --dry-run=client -o yaml | kubectl apply -f -

echo "Creating httpbin deployments"
kubectl create deployment httpbin --image=kennethreitz/httpbin -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -
kubectl create deployment httpbin2 --image=kennethreitz/httpbin -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -

kubectl create deployment httpbin --image=kennethreitz/httpbin -n mizu-tests2 --dry-run=client -o yaml | kubectl apply -f -

echo "Creating redis deployment"
kubectl create deployment redis --image=redis -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -

echo "Creating rabbitmq deployment"
kubectl create deployment rabbitmq --image=rabbitmq -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -

echo "Creating httpbin services"
kubectl expose deployment httpbin --type=NodePort --port=80 -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -
kubectl expose deployment httpbin2 --type=NodePort --port=80 -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -

kubectl expose deployment httpbin --type=NodePort --port=80 -n mizu-tests2 --dry-run=client -o yaml | kubectl apply -f -

echo "Creating redis service"
kubectl expose deployment redis --type=LoadBalancer --port=6379 -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -

echo "Creating rabbitmq service"
kubectl expose deployment rabbitmq --type=LoadBalancer --port=5672 -n mizu-tests --dry-run=client -o yaml | kubectl apply -f -

# TODO: need to understand how to fail if address already in use
echo "Starting proxy"
rm -f ${PROXY_LOG}
kubectl proxy --port=8080 > ${PROXY_LOG} &
PID1=$!
echo "kubectl proxy process id is ${PID1} and log of proxy in ${PROXY_LOG}"

if [[ -z "${CI}" ]]; then
  echo "Setting env var of mizu ci image"
  export MIZU_CI_IMAGE="mizu/ci:0.0"
  echo "Build agent image"
  docker build -t "${MIZU_CI_IMAGE}" .
else
  echo "not building docker image in CI because it is created as separate step"
fi

minikube image load "${MIZU_CI_IMAGE}"

echo "Build cli"
cd cli && make build GIT_BRANCH=ci SUFFIX=ci

# TODO: need to understand how to fail if password is asked (sudo)
echo "Starting tunnel"
rm -f ${TUNNEL_LOG}
minikube tunnel > ${TUNNEL_LOG} &
PID2=$!
echo "Minikube tunnel process id is ${PID2} and log of tunnel in ${TUNNEL_LOG}"
