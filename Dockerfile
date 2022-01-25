ARG ARCH=amd64

### Front-end
FROM node:16 AS front-end

WORKDIR /app/ui-build

COPY ui/package.json .
COPY ui/package-lock.json .
RUN npm i
COPY ui .
RUN npm run build



### Builder image for AMD64 architecture
FROM golang:1.16-alpine AS builder-amd64

ENV CGO_ENABLED=1 GOOS=linux
ENV GOARCH=amd64

RUN apk add libpcap-dev gcc g++ make bash perl-utils



### Builder image for ARM64 architecture
FROM dockcross/linux-arm64-musl AS builder-arm64v8

ENV CGO_ENABLED=1 GOOS=linux
ENV GOARCH=arm64 CGO_CFLAGS="-I/work/libpcap"

# Install Go
RUN curl https://go.dev/dl/go1.16.13.linux-amd64.tar.gz -Lo ./go.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go.linux-amd64.tar.gz
ENV PATH "$PATH:/usr/local/go/bin"

# Compile libpcap
RUN curl https://www.tcpdump.org/release/libpcap-1.10.1.tar.gz -Lo ./libpcap.tar.gz
RUN tar -xzf libpcap.tar.gz && mv ./libpcap-* ./libpcap
RUN cd ./libpcap && ./configure --host=arm && make
RUN cp /work/libpcap/libpcap.a /usr/xcc/aarch64-linux-musl-cross/lib/gcc/aarch64-linux-musl/*/



### Final builder image where the build happens
FROM builder-${ARCH} AS builder

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

ARG COMMIT_HASH
ARG GIT_BRANCH
ARG BUILD_TIMESTAMP
ARG SEM_VER=0.0.0

WORKDIR /app/agent-build

RUN go build -ldflags="-extldflags=-static \
    -X 'mizuserver/pkg/version.GitCommitHash=${COMMIT_HASH}' \
    -X 'mizuserver/pkg/version.Branch=${GIT_BRANCH}' \
    -X 'mizuserver/pkg/version.BuildTimestamp=${BUILD_TIMESTAMP}' \
    -X 'mizuserver/pkg/version.SemVer=${SEM_VER}'" -o mizuagent .

COPY devops/build_extensions.sh ..
RUN cd .. && /bin/bash build_extensions.sh



### The shipped image
ARG ARCH=amd64
FROM ${ARCH}/busybox:latest

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
