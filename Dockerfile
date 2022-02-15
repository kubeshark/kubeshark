ARG BUILDARCH=amd64
ARG TARGETARCH=amd64

### Front-end
FROM node:16 AS front-end

WORKDIR /app/ui-build

COPY ui/package.json .
COPY ui/package-lock.json .
RUN npm i
COPY ui .
RUN npm run build
RUN npm run build-ent

### Base builder image for native builds architecture
FROM golang:1.17-alpine AS builder-native-base
ENV CGO_ENABLED=1 GOOS=linux
RUN apk add libpcap-dev g++ perl-utils


### Intermediate builder image for x86-64 to x86-64 native builds
FROM builder-native-base AS builder-from-amd64-to-amd64
ENV GOARCH=amd64


### Intermediate builder image for AArch64 to AArch64 native builds
FROM builder-native-base AS builder-from-arm64v8-to-arm64v8
ENV GOARCH=arm64


### Builder image for x86-64 to AArch64 cross-compilation
FROM up9inc/linux-arm64-musl-go-libpcap AS builder-from-amd64-to-arm64v8
ENV CGO_ENABLED=1 GOOS=linux
ENV GOARCH=arm64 CGO_CFLAGS="-I/work/libpcap"


### Final builder image where the build happens
# Possible build strategies:
# BUILDARCH=amd64 TARGETARCH=amd64
# BUILDARCH=arm64v8 TARGETARCH=arm64v8
# BUILDARCH=amd64 TARGETARCH=arm64v8
ARG BUILDARCH=amd64
ARG TARGETARCH=amd64
FROM builder-from-${BUILDARCH}-to-${TARGETARCH} AS builder

# Move to agent working directory (/agent-build)
WORKDIR /app/agent-build

COPY agent/go.mod agent/go.sum ./
COPY shared/go.mod shared/go.mod ../shared/
COPY tap/go.mod tap/go.mod ../tap/
COPY tap/api/go.mod ../tap/api/
COPY tap/extensions/amqp/go.mod ../tap/extensions/amqp/
COPY tap/extensions/http/go.mod ../tap/extensions/http/
COPY tap/extensions/kafka/go.mod ../tap/extensions/kafka/
COPY tap/extensions/redis/go.mod ../tap/extensions/redis/
RUN go mod download
# cheap trick to make the build faster (as long as go.mod did not change)
RUN go list -f '{{.Path}}@{{.Version}}' -m all | sed 1d | grep -e 'go-cache' | xargs go get

# Copy and build agent code
COPY shared ../shared
COPY tap ../tap
COPY agent .

ARG COMMIT_HASH
ARG GIT_BRANCH
ARG BUILD_TIMESTAMP
ARG VER=0.0

WORKDIR /app/agent-build

RUN go build -ldflags="-extldflags=-static -s -w \
    -X 'github.com/up9inc/mizu/agent/pkg/version.GitCommitHash=${COMMIT_HASH}' \
    -X 'github.com/up9inc/mizu/agent/pkg/version.Branch=${GIT_BRANCH}' \
    -X 'github.com/up9inc/mizu/agent/pkg/version.BuildTimestamp=${BUILD_TIMESTAMP}' \
    -X 'github.com/up9inc/mizu/agent/pkg/version.Ver=${VER}'" -o mizuagent .

# Download Basenine executable, verify the sha1sum
ADD https://github.com/up9inc/basenine/releases/download/v0.4.16/basenine_linux_${GOARCH} ./basenine_linux_${GOARCH}
ADD https://github.com/up9inc/basenine/releases/download/v0.4.16/basenine_linux_${GOARCH}.sha256 ./basenine_linux_${GOARCH}.sha256
RUN shasum -a 256 -c basenine_linux_${GOARCH}.sha256
RUN chmod +x ./basenine_linux_${GOARCH}
RUN mv ./basenine_linux_${GOARCH} ./basenine


### The shipped image
ARG TARGETARCH=amd64
FROM ${TARGETARCH}/busybox:latest

# gin-gonic runs in debug mode without this
ENV GIN_MODE=release

WORKDIR /app/data/
WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/agent-build/mizuagent", "."]
COPY --from=builder ["/app/agent-build/basenine", "/usr/local/bin/basenine"]
COPY --from=front-end ["/app/ui-build/build", "site"]
COPY --from=front-end ["/app/ui-build/build-ent", "site-standalone"]

# this script runs both apiserver and passivetapper and exits either if one of them exits, preventing a scenario where the container runs without one process
ENTRYPOINT ["/app/mizuagent"]
