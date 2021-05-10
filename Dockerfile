FROM node:14-slim AS site-build

WORKDIR /ui-build

COPY ui .
RUN npm i
RUN npm run build


FROM golang:1.16-alpine AS builder
# Set necessary environment variables needed for our image.
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

RUN apk add libpcap-dev gcc g++ make

# Move to api working directory (/api-build).
WORKDIR /api-build

COPY api/go.mod api/go.sum ./
RUN go mod download
# cheap trick to make the build faster (As long as go.mod wasn't changes)
RUN go list -f '{{.Path}}@{{.Version}}' -m all | sed 1d | grep -e 'go-cache' -e 'sqlite' | xargs go get

# Copy and build api code
COPY api .
RUN go build -ldflags="-s -w" -o mizuagent .


FROM alpine:3.13.5

RUN apk add bash libpcap-dev tcpdump
WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/api-build/mizuagent", "."]
COPY --from=site-build ["/ui-build/build", "site"]

COPY api/start.sh .

# this script runs both apiserver and passivetapper and exits either if one of them exits, preventing a scenario where the container runs without one process
CMD "./start.sh"
