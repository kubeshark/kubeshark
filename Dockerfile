FROM node:14-slim AS site-build

WORKDIR /app/ui-build

COPY ui/package.json .
COPY ui/package-lock.json .
RUN npm i
COPY ui .
RUN npm run build


FROM golang:1.16-alpine AS builder
# Set necessary environment variables needed for our image.
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

RUN apk add libpcap-dev gcc g++ make bash perl-utils

# Move to agent working directory (/agent-build).
WORKDIR /app/agent-build

COPY agent/go.mod agent/go.sum ./
COPY shared/go.mod shared/go.mod ../shared/
COPY tap/go.mod tap/go.mod ../tap/
COPY tap/api/go.* ../tap/api/
RUN go mod download
# cheap trick to make the build faster (As long as go.mod wasn't changes)
RUN go list -f '{{.Path}}@{{.Version}}' -m all | sed 1d | grep -e 'go-cache' -e 'sqlite' | xargs go get

ARG COMMIT_HASH
ARG GIT_BRANCH
ARG BUILD_TIMESTAMP
ARG SEM_VER

# Copy and build agent code
COPY shared ../shared
COPY tap ../tap
COPY agent .
RUN go build -ldflags="-s -w \
     -X 'mizuserver/pkg/version.GitCommitHash=${COMMIT_HASH}' \
     -X 'mizuserver/pkg/version.Branch=${GIT_BRANCH}' \
     -X 'mizuserver/pkg/version.BuildTimestamp=${BUILD_TIMESTAMP}' \
     -X 'mizuserver/pkg/version.SemVer=${SEM_VER}'" -o mizuagent .

# Download Basenine executable, verify the sha1sum and move it to a directory in $PATH
ADD https://github.com/up9inc/basenine/releases/download/v0.2.1/basenine_linux_amd64 ./basenine_linux_amd64
ADD https://github.com/up9inc/basenine/releases/download/v0.2.1/basenine_linux_amd64.sha256 ./basenine_linux_amd64.sha256
RUN shasum -a 256 -c basenine_linux_amd64.sha256
RUN chmod +x ./basenine_linux_amd64

COPY devops/build_extensions.sh ..
RUN cd .. && /bin/bash build_extensions.sh

FROM alpine:3.14

RUN apk add bash libpcap-dev tcpdump

WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/agent-build/mizuagent", "."]
COPY --from=builder ["/app/agent-build/basenine_linux_amd64", "/usr/local/bin/basenine"]
COPY --from=builder ["/app/agent/build/extensions", "extensions"]
COPY --from=site-build ["/app/ui-build/build", "site"]

# gin-gonic runs in debug mode without this
ENV GIN_MODE=release

# this script runs both apiserver and passivetapper and exits either if one of them exits, preventing a scenario where the container runs without one process
ENTRYPOINT "/app/mizuagent"
