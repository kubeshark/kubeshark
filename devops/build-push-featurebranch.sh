#!/bin/bash
set -e

GCP_PROJECT=up9-docker-hub
REPOSITORY=gcr.io/$GCP_PROJECT
SERVER_NAME=mizu
GIT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2 | tr '[:upper:]' '[:lower:]')

DOCKER_REPO=$REPOSITORY/$SERVER_NAME/$GIT_BRANCH
VER=${VER=0.0}

DOCKER_TAGGED_BUILDS=("$DOCKER_REPO:latest" "$DOCKER_REPO:$VER")

if [ "$GIT_BRANCH" = 'develop' -o "$GIT_BRANCH" = 'master' -o "$GIT_BRANCH" = 'main' ]
then
  echo "Pushing to $GIT_BRANCH is allowed only via CI"
  exit 1
fi

echo "building ${DOCKER_TAGGED_BUILDS[@]}"
DOCKER_TAGS_ARGS=$(echo ${DOCKER_TAGGED_BUILDS[@]/#/-t }) # "-t FIRST_TAG -t SECOND_TAG ..."
docker build $DOCKER_TAGS_ARGS --build-arg VER=${VER} --build-arg BUILD_TIMESTAMP=${BUILD_TIMESTAMP} --build-arg GIT_BRANCH=${GIT_BRANCH} --build-arg COMMIT_HASH=${COMMIT_HASH} .

for DOCKER_TAG in "${DOCKER_TAGGED_BUILDS[@]}"
do
  echo pushing "$DOCKER_TAG"
  docker push "$DOCKER_TAG"
done
