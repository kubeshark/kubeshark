<p align="center">
  <img src="assets/kubeshark-logo.svg" alt="Kubeshark: Traffic viewer for Kubernetes." height="128px"/>
</p>

<p align="center">
    <a href="https://github.com/kubeshark/kubeshark/blob/main/LICENSE">
        <img alt="GitHub License" src="https://img.shields.io/github/license/kubeshark/kubeshark?logo=GitHub&style=flat-square">
    </a>
    <a href="https://github.com/kubeshark/kubeshark/releases/latest">
        <img alt="GitHub Latest Release" src="https://img.shields.io/github/v/release/kubeshark/kubeshark?logo=GitHub&style=flat-square">
    </a>
    <a href="https://hub.docker.com/r/kubeshark/kubeshark">
      <img alt="Docker pulls" src="https://img.shields.io/docker/pulls/kubeshark/kubeshark?color=%23099cec&logo=Docker&style=flat-square">
    </a>
    <a href="https://hub.docker.com/r/kubeshark/kubeshark">
      <img alt="Image size" src="https://img.shields.io/docker/image-size/kubeshark/kubeshark/latest?logo=Docker&style=flat-square">
    </a>
		<a href="https://discord.gg/2ABWCpDCjt">
      <img alt="Discord" src="https://img.shields.io/discord/1042559155224973352?logo=Discord&style=flat-square&label=discord">
    </a>
    <a href="https://join.slack.com/t/kubeshark/shared_invite/zt-1k3sybpq9-uAhFkuPJiJftKniqrGHGhg">
      <img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-white?logo=Slack&style=flat-square&label=slack">
    </a>
</p>

<p>
<p align="center">
Mizu (by UP9) is now Kubeshark, read more about it <a href="https://www.kubeshark.co/mizu-is-now-kubeshark">here</a>.
</p>

Kubeshark, the API Traffic Viewer for kubernetes, provides deep visibility and monitoring of all API traffic and payloads going in, out and across containers and pods inside a Kubernetes cluster.

Think of a combination of Chrome Dev Tools, TCPDump and Wireshark, re-invented for Kubernetes.

![Simple UI](assets/kubeshark-ui.png)

## Download

Kubeshark uses a ~45MB pre-compiled executable binary to communicate with the Kubernetes API. We recommend downloading the `kubeshark` CLI by using one of these options:

> Choose the right binary, download and use directly from [the releases section](https://github.com/kubeshark/kubeshark/releases/); or

> Use this :point_down: shell script to download the right binary for your operating system and CPU architecture.

```shell
sh <(curl -Ls https://kubeshark.co/install)
```

> Compile from source

## Run

Use the `kubeshark` CLI to capture and view streaming API traffic in real time.

```shell
kubeshark tap
```
### Troubleshooting Installation
If something doesn't work or simply to play it safe prior to installing, make sure that:

> Make sure you have access to https://hub.docker.com/

> Make sure `kubeshark` executable in your `PATH`.

### Select Pods

#### Monitoring a Specific Pod:

```shell
kubeshark tap catalogue-b87b45784-sxc8q
```

#### Monitoring a Set of Pods Using Regex:

```shell
kubeshark tap "(catalo*|front-end*)"
```

### Specify the Namespace

By default, Kubeshark is deployed into the `default` namespace.
To specify a different namespace:

```
kubeshark tap -n sock-shop
```

### Specify All Namespaces

The default deployment strategy of Kubeshark waits for the new pods
to be created. To simply deploy to all existing namespaces run:

```
kubeshark tap -A
```

## Documentation

Visit our documentation website: [docs.kubeshark.co](https://docs.kubeshark.co)

The documentation resources are open-source and can be found on GitHub: [kubeshark/docs](https://github.com/kubeshark/docs)

## Contributing

We ❤️ pull requests! See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for the contribution guide.

## Code of Conduct

This project is for everyone. We ask that our users and contributors take a few minutes to review our [Code of Conduct](docs/CODE_OF_CONDUCT.md).
