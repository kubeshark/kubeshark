# mizu agent
Agent for MIZU (API server and tapper)
Basic APIs:
* /stats - retrieve statistics of collected data
* /viewer - web ui

## Remote Debugging

### Setup remote debugging
1. Run `go get github.com/go-delve/delve/cmd/dlv`
2. Create a "Go Remote" run/debug configuration in Intellij, set to localhost:2345
3. Build and push a debug image using
   `docker build . -t  gcr.io/up9-docker-hub/mizu/debug:latest --build-arg ARCH=amd64 --build-arg GOARCH=amd64 -f debug.Dockerfile && docker push gcr.io/up9-docker-hub/mizu/debug:latest`

### Connecting
1. Start mizu using the cli with the debug
   image `mizu tap --set agent-image=gcr.io/up9-docker-hub/mizu/debug:latest {tapped_pod_name}`
2. Forward the debug port using `kubectl port-forward -n default mizu-api-server 2345:2345`
3. Run the run/debug configuration you've created earlier in Intellij.

<small>Do note that dlv won't start the api until a debugger connects to it.</small>
