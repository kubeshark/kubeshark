# creates image in which mizu agent is remotely debuggable using delve
FROM node:14-slim AS site-build

WORKDIR /app/ui-build

COPY ui .
RUN npm i
RUN npm run build


FROM golang:1.16-alpine AS builder
# Set necessary environment variables needed for our image.
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

RUN apk add libpcap-dev gcc g++ make bash

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
ARG SEM_VER=0.0.0

# Copy and build agent code
COPY shared ../shared
COPY tap ../tap
COPY agent .
RUN go build -gcflags="all=-N -l" -o mizuagent .

COPY devops/build_extensions_debug.sh ..
RUN cd .. && /bin/bash build_extensions_debug.sh


FROM golang:1.16-alpine

RUN apk add bash libpcap-dev tcpdump
WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/agent-build/mizuagent", "."]
COPY --from=builder ["/app/agent/build/extensions", "extensions"]
COPY --from=site-build ["/app/ui-build/build", "site"]

# install remote debugging tool
RUN go get github.com/go-delve/delve/cmd/dlv

ENTRYPOINT "/app/mizuagent"
#CMD ["sh", "-c", "dlv --headless=true --listen=:2345 --log --api-version=2 --accept-multiclient exec ./mizuagent -- --api-server"]
