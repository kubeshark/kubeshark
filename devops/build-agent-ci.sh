#!/bin/bash
set -e

GCP_PROJECT=up9-docker-hub
REPOSITORY=gcr.io/$GCP_PROJECT
SERVER_NAME=mizu
GIT_BRANCH=ci

DOCKER_REPO=$REPOSITORY/$SERVER_NAME/$GIT_BRANCH
VER=${VER=0.0}

DOCKER_TAGGED_BUILD="$DOCKER_REPO:$VER"

echo "pulling agent docker images"
docker pull gcr.io/up9-docker-hub/mizu/develop:28.0-dev14-arm64v8
docker pull gcr.io/up9-docker-hub/mizu/develop:28.0-dev14-amd64

echo "building $DOCKER_TAGGED_BUILD"
docker build -t ${DOCKER_TAGGED_BUILD} --build-arg VER=${VER} --build-arg BUILD_TIMESTAMP=${BUILD_TIMESTAMP} --build-arg GIT_BRANCH=${GIT_BRANCH} --build-arg COMMIT_HASH=${COMMIT_HASH} .
