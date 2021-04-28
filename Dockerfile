FROM node:14-slim AS site-build

WORKDIR /ui-build

COPY ui ./
RUN npm i
RUN npm run build


FROM golang:1.16-alpine AS builder
# Set necessary environment variables needed for our image.
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

RUN apk add libpcap-dev gcc g++ make

# Move to tapper working directory (/tap-build).
WORKDIR /tap-build
COPY tap/go.mod tap/go.sum ./
RUN go mod download
# Copy and build tapper code
COPY tap/src ./
RUN go build -ldflags="-s -w" -o passivetapper .

# Move to api working directory (/api-build).
WORKDIR /api-build
COPY api/go.mod api/go.sum ./
RUN go mod download
# Copy and build api code
COPY api .
RUN go build -ldflags="-s -w" -o apiserver .


FROM alpine:3.13.5
RUN apk add parallel libpcap-dev
RUN apk add tcpdump
# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/api-build/apiserver", "/"]
COPY --from=builder ["/tap-build/passivetapper", "/"]
COPY --from=site-build ["/ui-build/build", "/site"]

# parallel will exit if one of the executables exits, ensuring this container does not run without one of the processes

ENV HOST_MODE="1"
CMD parallel --halt now,done=1 ::: './apiserver' './passivetapper -i any -hardump -hardir /tmp/up9hars -harentriesperfile 10'