# mizu CLI
## Usage
`./main {pod_name_regex}`

### Optional Flags

| flag                 | default          | purpose                                                                                                      |
|----------------------|------------------|--------------------------------------------------------------------------------------------------------------|
| `--no-dashboard`     | `false`          | Don't host the dashboard (not applicable at the moment)                                                      |
| `--dashboard-port`   | `3000`           | local port that dashboard will be forwarded to                                                               |
| `--namespace`        |                  | use namespace different than the one found in kubeconfig                                                     |
| `--kubeconfig`       |                  | Path to custom kubeconfig file                                                                               |

There are some extra flags defined in code that will show up in `./main --help`, these are non functional stubs for now

## Installation
Make sure you have go v1.16 installed.
1. cd to `mizu/cmd`
2. Run `go mod download` (may take a moment)
3. Run `go build main.go`

Alternatively, you can build+run directly using `go run main.go {pod_name_regex}`


## Known issues
* mid-flight port forwarding failures are not detected and no indication will be shown when this occurs
