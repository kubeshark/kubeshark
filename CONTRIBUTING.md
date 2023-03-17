![Kubeshark: The API Traffic Analyzer for Kubernetes](https://raw.githubusercontent.com/kubeshark/assets/master/svg/kubeshark-logo.svg)

# Contributing to Kubeshark

We welcome code contributions from the community.
Please read and follow the guidelines below.

## Communication

* Before starting work on a major feature, please reach out to us via [GitHub](https://github.com/kubeshark/kubeshark), [Discord](https://discord.gg/WkvRGMUcx7), [Slack](https://join.slack.com/t/kubeshark/shared_invite/zt-1k3sybpq9-uAhFkuPJiJftKniqrGHGhg), [email](mailto:info@kubeshark.co), etc. We will make sure no one else is already working on it. A _major feature_ is defined as any change that is > 100 LOC altered (not including tests), or changes any user-facing behavior
* Small patches and bug fixes don't need prior communication.

## Contribution Requirements

* Code style - most of the code is written in Go, please follow [these guidelines](https://golang.org/doc/effective_go)
* Go-tools compatible (`go get`, `go test`, etc.)
* Code coverage for unit tests must not decrease.
* Code must be usefully commented. Not only for developers on the project, but also for external users of these packages
* When reviewing PRs, you are encouraged to use Golang's [code review comments page](https://github.com/golang/go/wiki/CodeReviewComments)
* Project follows [Google JSON Style Guide](https://google.github.io/styleguide/jsoncstyleguide.xml) for the REST APIs that are provided.
