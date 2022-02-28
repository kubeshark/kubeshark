#!/bin/bash

# Create a user in Minikube cluster "minikube"
# Create context for user
# Usage:
#  ./create_user.sh <username>

set -e

NEW_USERNAME=$1
CERT_DIR="${HOME}/certs"
KEY_FILE="${CERT_DIR}/${NEW_USERNAME}.key"
CRT_FILE="${CERT_DIR}/${NEW_USERNAME}.crt"
MINIKUBE_KEY_FILE="${HOME}/.minikube/ca.key"
MINIKUBE_CRT_FILE="${HOME}/.minikube/ca.crt"
DAYS=1

echo "Creating user and context for username \"${NEW_USERNAME}\" in Minikube cluster"

if ! command -v openssl &> /dev/null
then
	 echo "Installing openssl"
	 sudo apt-get update
	 sudo apt-get install openssl
fi

echo "Creating certificate for user \"${NEW_USERNAME}\""
mkdir -p ${CERT_DIR}
echo "Generating key \"${KEY_FILE}\""
openssl genrsa -out "${KEY_FILE}" 2048
echo "Generating crt \"${CRT_FILE}\""
openssl req -new -key "${KEY_FILE}" -out "${CRT_FILE}" -subj "/CN=${NEW_USERNAME}/O=group1"
openssl x509 -req -in "${CRT_FILE}" -CA "${MINIKUBE_CRT_FILE}" -CAkey "${MINIKUBE_KEY_FILE}" -CAcreateserial -out "${CRT_FILE}" -days $DAYS

echo "Creating context for user \"${NEW_USERNAME}\""
kubectl config set-credentials "${NEW_USERNAME}" --client-certificate="${CRT_FILE}" --client-key="${KEY_FILE}"
kubectl config set-context "${NEW_USERNAME}" --cluster=minikube --user="${NEW_USERNAME}"
