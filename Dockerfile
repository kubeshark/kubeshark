ARG TARGETARCH=amd64

### Base builder image for native builds architecture
FROM golang:1.17-alpine AS builder-native-base
ENV CGO_ENABLED=1 GOOS=linux
RUN apk add --no-cache g++ perl-utils


### Intermediate builder image for x86-64 native builds
FROM builder-native-base AS builder-for-amd64
ENV GOARCH=amd64
ENV BPF_TARGET=amd64 BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_x86"


### Intermediate builder image for AArch64 native builds
FROM builder-native-base AS builder-for-arm64v8
ENV GOARCH=arm64
ENV BPF_TARGET=arm64 BPF_CFLAGS="-O2 -g -D__TARGET_ARCH_arm64"


### Final builder image where the build happens
# Possible build strategies:
# TARGETARCH=amd64
# TARGETARCH=arm64v8
ARG TARGETARCH=amd64
FROM builder-for-${TARGETARCH} AS builder

# Move to agent working directory (/agent-build)
WORKDIR /app/agent-build

COPY agent/go.mod agent/go.sum ./
COPY shared/go.mod shared/go.mod ../shared/
COPY logger/go.mod logger/go.mod ../logger/
RUN go mod download

# Copy and build agent code
COPY shared ../shared
COPY logger ../logger
COPY agent .

ARG COMMIT_HASH
ARG GIT_BRANCH
ARG BUILD_TIMESTAMP
ARG VER=0.0

WORKDIR /app/agent-build

RUN go build -ldflags="-extldflags=-static -s -w \
    -X 'github.com/kubeshark/kubeshark/agent/pkg/version.GitCommitHash=${COMMIT_HASH}' \
    -X 'github.com/kubeshark/kubeshark/agent/pkg/version.Branch=${GIT_BRANCH}' \
    -X 'github.com/kubeshark/kubeshark/agent/pkg/version.BuildTimestamp=${BUILD_TIMESTAMP}' \
    -X 'github.com/kubeshark/kubeshark/agent/pkg/version.Ver=${VER}'" -o kubesharkagent .

# Download Basenine executable, verify the sha1sum
ADD https://github.com/up9inc/basenine/releases/download/v0.8.3/basenine_linux_${GOARCH} ./basenine_linux_${GOARCH}
ADD https://github.com/up9inc/basenine/releases/download/v0.8.3/basenine_linux_${GOARCH}.sha256 ./basenine_linux_${GOARCH}.sha256

RUN shasum -a 256 -c basenine_linux_"${GOARCH}".sha256 && \
    chmod +x ./basenine_linux_"${GOARCH}" && \
    mv ./basenine_linux_"${GOARCH}" ./basenine

### The shipped image
ARG TARGETARCH=amd64
FROM ${TARGETARCH}/busybox:latest
# gin-gonic runs in debug mode without this
ENV GIN_MODE=release

WORKDIR /app/data/
WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/agent-build/kubesharkagent", "."]
COPY --from=builder ["/app/agent-build/basenine", "/usr/local/bin/basenine"]

ENTRYPOINT ["/app/kubesharkagent"]
