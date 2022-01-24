ARG ARCH=amd64

### Front-end
FROM node:16 AS front-end

WORKDIR /app/ui-build

COPY ui/package.json .
COPY ui/package-lock.json .
RUN npm i
COPY ui .
RUN npm run build


### Base of the builder image
FROM golang:1.16-bullseye AS builder-base

# Set necessary environment variables needed for our image.
ENV CGO_ENABLED=1 GOOS=linux GOARCH=${GOARCH}

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        build-essential \
        libpcap-dev

# Move to agent working directory (/agent-build).
WORKDIR /app/agent-build

COPY agent/go.mod agent/go.sum ./
COPY shared/go.mod shared/go.mod ../shared/
COPY tap/go.mod tap/go.mod ../tap/
COPY tap/api/go.* ../tap/api/
RUN go mod download
# cheap trick to make the build faster (as long as go.mod did not change)
RUN go list -f '{{.Path}}@{{.Version}}' -m all | sed 1d | grep -e 'go-cache' | xargs go get

# Copy and build agent code
COPY shared ../shared
COPY tap ../tap
COPY agent .


### Intermediate builder image for AMD64 architecture
FROM builder-base AS builder-amd64

ENV GOARCH=amd64


### Intermediate builder image for ARM64 architecture
FROM builder-base AS builder-arm64v8

ENV GOARCH=arm64
ENV CC=aarch64-linux-gnu-gcc
ENV PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig

RUN dpkg --add-architecture arm64 \
    && apt-get update \
    && apt-get install -y --no-install-recommends \
        gcc-aarch64-linux-gnu \
        libpcap-dev:arm64


### Final builder image where the building happens
FROM builder-${ARCH} AS builder

ARG COMMIT_HASH
ARG GIT_BRANCH
ARG BUILD_TIMESTAMP
ARG SEM_VER=0.0.0

WORKDIR /app/agent-build

RUN go build -ldflags="-extldflags '-fuse-ld=bfd' -s -w \
    -X 'mizuserver/pkg/version.GitCommitHash=${COMMIT_HASH}' \
    -X 'mizuserver/pkg/version.Branch=${GIT_BRANCH}' \
    -X 'mizuserver/pkg/version.BuildTimestamp=${BUILD_TIMESTAMP}' \
    -X 'mizuserver/pkg/version.SemVer=${SEM_VER}'" -o mizuagent .

COPY devops/build_extensions.sh ..
RUN cd .. && /bin/bash build_extensions.sh



### The shipped image
ARG ARCH=amd64
FROM up9inc/debian-pcap:stable-slim-${ARCH}

WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/agent-build/mizuagent", "."]
COPY --from=builder ["/app/agent/build/extensions", "extensions"]
COPY --from=front-end ["/app/ui-build/build", "site"]
COPY --from=front-end ["/app/ui-build/build-ent", "site-standalone"]

# gin-gonic runs in debug mode without this
ENV GIN_MODE=release

# this script runs both apiserver and passivetapper and exits either if one of them exits, preventing a scenario where the container runs without one process
ENTRYPOINT "/app/mizuagent"
