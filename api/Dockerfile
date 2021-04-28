FROM golang:1.16 AS builder

# Move to working directory (/site).
WORKDIR /build

# Copy and download dependency using go mod.
COPY go.mod go.sum ./
RUN go mod download

# Copy the code into the container.
COPY . .

# Set necessary environmet variables needed for our image and site the API server.
ENV GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o apiserver .

FROM golang:1.16

# Copy binary and config files from /site to root folder of scratch container.
COPY --from=builder ["/build/apiserver", "/"]

# Export necessary port.
EXPOSE 8899

COPY site /go/site
COPY entries.db /go/entries.db

# Command to run when starting the container.
ENTRYPOINT ["/apiserver"]
