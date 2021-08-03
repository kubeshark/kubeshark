#!/bin/bash
set -e

SERVER_NAME=mizu
GCP_PROJECT=up9-docker-hub
REPOSITORY=gcr.io/$GCP_PROJECT
GIT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2 | tr '[:upper:]' '[:lower:]')
SEM_VER=${SEM_VER=0.0.0}
DOCKER_REPO=$REPOSITORY/$SERVER_NAME/$GIT_BRANCH
DOCKER_TAGGED_BUILDS=("$DOCKER_REPO:latest" "$DOCKER_REPO:$SEM_VER")

if [ "$GIT_BRANCH" = 'develop' -o "$GIT_BRANCH" = 'master' -o "$GIT_BRANCH" = 'main' ]
then
  echo "Pushing to $GIT_BRANCH is allowed only via CI"
  exit 1
fi

for DOCKER_TAG in ${DOCKER_TAGGED_BUILDS[@]}
do
        echo "building $DOCKER_TAG"
        docker build -t "$DOCKER_TAG" --build-arg SEM_VER=${SEM_VER} --build-arg BUILD_TIMESTAMP=${BUILD_TIMESTAMP} --build-arg GIT_BRANCH=${GIT_BRANCH} --build-arg COMMIT_HASH=${COMMIT_HASH} .
done

for DOCKER_TAG in ${DOCKER_TAGGED_BUILDS[@]}
do
        echo pushing to "$REPOSITORY"
        docker push "$DOCKER_TAG"
done
