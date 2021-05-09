#!/bin/bash
set -e

SERVER_NAME=mizu
GCP_PROJECT=up9-docker-hub
REPOSITORY=gcr.io/$GCP_PROJECT
GIT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2 | tr '[:upper:]' '[:lower:]')
DOCKER_TAGGED_BUILD=$REPOSITORY/$SERVER_NAME/$GIT_BRANCH:latest

if [ "$GIT_BRANCH" = 'develop' -o "$GIT_BRANCH" = 'master' -o "$GIT_BRANCH" = 'main' ]
then
  echo "Pushing to $GIT_BRANCH is allowed only via CI"
  exit 1
fi

echo "building $DOCKER_TAGGED_BUILD"
docker build -t "$DOCKER_TAGGED_BUILD" .

echo pushing to "$REPOSITORY"
docker push "$DOCKER_TAGGED_BUILD"
