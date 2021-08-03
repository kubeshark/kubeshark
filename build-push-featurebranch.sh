#!/bin/bash
set -e

SERVER_NAME=mizu
GCP_PROJECT=up9-docker-hub
REPOSITORY=gcr.io/$GCP_PROJECT
GIT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2 | tr '[:upper:]' '[:lower:]')
SEM_VER=${SEM_VER=0.0.0}
DOCKER_TAGGED_BUILD=$REPOSITORY/$SERVER_NAME/$GIT_BRANCH:latest

if [ "$GIT_BRANCH" = 'develop' -o "$GIT_BRANCH" = 'master' -o "$GIT_BRANCH" = 'main' ]
then
  echo "Pushing to $GIT_BRANCH is allowed only via CI"
  exit 1
fi

echo "building $DOCKER_TAGGED_BUILD"
docker build -t "$DOCKER_TAGGED_BUILD" --build-arg SEM_VER=${SEM_VER} --build-arg BUILD_TIMESTAMP=${BUILD_TIMESTAMP} --build-arg GIT_BRANCH=${GIT_BRANCH} --build-arg COMMIT_HASH=${COMMIT_HASH} .

echo pushing to "$REPOSITORY"
docker push "$DOCKER_TAGGED_BUILD"
