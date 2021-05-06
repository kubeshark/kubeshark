# mizu CLI
## Usage
`./mizu {pod_name_regex}`

### Optional Flags

| flag                 | default          | purpose                                                                                                      |
|----------------------|------------------|--------------------------------------------------------------------------------------------------------------|
| `--no-gui`           | `false`          | Don't host the web interface (not applicable at the moment)                                                      |
| `--gui-port`         | `8899`           | local port that web interface will be forwarded to                                                               |
| `--namespace`        |                  | use namespace different than the one found in kubeconfig                                                     |
| `--kubeconfig`       |                  | Path to custom kubeconfig file                                                                               |

There are some extra flags defined in code that will show up in `./mizu --help`, these are non functional stubs for now

## Installation
Make sure your go version is at least 1.11
1. cd to `mizu/cli`
2. Run `go mod download` (may take a moment)
3. Run `go build mizu.go`

Alternatively, you can build+run directly using `go run mizu.go {pod_name_regex}`


## Known issues
* mid-flight port forwarding failures are not detected and no indication will be shown when this occurs
