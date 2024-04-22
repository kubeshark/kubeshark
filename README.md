<p align="center">
  <img src="https://raw.githubusercontent.com/kubeshark/assets/master/svg/kubeshark-logo.svg" alt="Kubeshark: Traffic analyzer for Kubernetes." height="128px"/>
</p>

<p align="center">
    <a href="https://github.com/kubeshark/kubeshark/releases/latest">
        <img alt="GitHub Latest Release" src="https://img.shields.io/github/v/release/kubeshark/kubeshark?logo=GitHub&style=flat-square">
    </a>
    <a href="https://hub.docker.com/r/kubeshark/worker">
      <img alt="Docker pulls" src="https://img.shields.io/docker/pulls/kubeshark/worker?color=%23099cec&logo=Docker&style=flat-square">
    </a>
    <a href="https://hub.docker.com/r/kubeshark/worker">
      <img alt="Image size" src="https://img.shields.io/docker/image-size/kubeshark/kubeshark/latest?logo=Docker&style=flat-square">
    </a>
		<a href="https://discord.gg/WkvRGMUcx7">
      <img alt="Discord" src="https://img.shields.io/discord/1042559155224973352?logo=Discord&style=flat-square&label=discord">
    </a>
    <a href="https://join.slack.com/t/kubeshark/shared_invite/zt-1m90td3n7-VHxN_~V5kVp80SfQW3SfpA">
      <img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-green?logo=Slack&style=flat-square&label=slack">
    </a>
</p>

<p align="center">
  <b>
	  Want to see Kubeshark in action,  right now? Visit this
	  <a href="https://demo.kubeshark.co/">live demo deployment</a> of Kubeshark.
  </b>
</p>

**Kubeshark** is an API Traffic Analyzer for [**Kubernetes**](https://kubernetes.io/) providing real-time, protocol-level visibility into Kubernetes’ internal network, capturing and monitoring all traffic and payloads going in, out and across containers, pods, nodes and clusters.

![Simple UI](https://github.com/kubeshark/assets/raw/master/png/kubeshark-ui.png)

Think [TCPDump](https://en.wikipedia.org/wiki/Tcpdump) and [Wireshark](https://www.wireshark.org/) re-invented for Kubernetes

## Getting Started

Download **Kubeshark**'s binary distribution [latest release](https://github.com/kubeshark/kubeshark/releases/latest) and run following one of these examples:

```shell
kubeshark tap
```

```shell
kubeshark tap -n sock-shop "(catalo*|front-end*)"
```

Running any of the :point_up: above commands will open the [Web UI](https://docs.kubeshark.co/en/ui) in your browser which streams the traffic in your Kubernetes cluster in real-time.

### Homebrew

[Homebrew](https://brew.sh/) :beer: users install Kubeshark CLI with:

```shell
brew install kubeshark
```

### Helm

Add the helm repository and install the chart:

```shell
helm repo add kubeshark https://helm.kubeshark.co
‍helm install kubeshark kubeshark/kubeshark
```

## Building From Source

Clone this repository and run `make` command to build it. After the build is complete, the executable can be found at `./bin/kubeshark__`.

## Documentation

To learn more, read the [documentation](https://docs.kubeshark.co).

## Contributing

We :heart: pull requests! See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution guide.

## Code of Conduct

This project is for everyone. We ask that our users and contributors take a few minutes to review our [Code of Conduct](CODE_OF_CONDUCT.md).
