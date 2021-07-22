# creates image in which mizu api is remotely debuggable using delve
FROM node:14-slim AS site-build

WORKDIR /app/ui-build

COPY ui .
RUN npm i
RUN npm run build


FROM golang:1.16-alpine AS builder
# Set necessary environment variables needed for our image.
ENV CGO_ENABLED=1 GOOS=linux GOARCH=amd64

RUN apk add libpcap-dev gcc g++ make

# Move to api working directory (/api-build).
WORKDIR /app/api-build

COPY api/go.mod api/go.sum ./
COPY shared/go.mod shared/go.mod ../shared/
COPY tap/go.mod tap/go.mod ../tap/

RUN go mod download
# cheap trick to make the build faster (As long as go.mod wasn't changes)
RUN go list -f '{{.Path}}@{{.Version}}' -m all | sed 1d | grep -e 'go-cache' -e 'sqlite' | xargs go get

# Copy and build api code
COPY shared ../shared
COPY tap ../tap
COPY api .
RUN go build -gcflags="all=-N -l" -o mizuagent .


FROM golang:1.16-alpine

RUN apk add bash libpcap-dev tcpdump
WORKDIR /app

# Copy binary and config files from /build to root folder of scratch container.
COPY --from=builder ["/app/api-build/mizuagent", "."]
COPY --from=site-build ["/app/ui-build/build", "site"]

# install remote debugging tool
RUN go get github.com/go-delve/delve/cmd/dlv

ENTRYPOINT "/app/mizuagent"
#CMD ["sh", "-c", "dlv --headless=true --listen=:2345 --log --api-version=2 --accept-multiclient exec ./mizuagent -- --api-server"]
