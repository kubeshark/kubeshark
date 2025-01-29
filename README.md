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

**Kubeshark** is a network observability platform for [**Kubernetes**](https://kubernetes.io/), providing real-time, protocol-level visibility into Kubernetes’ network, enabling users to inspect all internal and external cluster connections, API calls, and data in transit. In addition, Kubeshark enables users to detect suspicious network behaviors, trigger automated actions, and gain unlimited insights into the network, also by using the latest GenAI technology.


![Simple UI](https://github.com/kubeshark/assets/raw/master/png/kubeshark-ui.png)

Think [TCPDump](https://en.wikipedia.org/wiki/Tcpdump) and [Wireshark](https://www.wireshark.org/) re-invented for Kubernetes

## Getting Started

Run `brew install kubeshark` or download **Kubeshark**'s binary distribution [latest release](https://github.com/kubeshark/kubeshark/releases/latest) and run `kubeshark tap`.
``

The [Web UI](https://docs.kubeshark.co/en/ui) should open in your browser and present a real-time view of all of your cluster's traffic.

### Homebrew

[Homebrew](https://brew.sh/) :beer: users install Kubeshark CLI with:

```shell
brew install kubeshark
```

Clean up:
```shell
kubeshark clean
```

### Helm

Add the helm repository and install the chart:

```shell
helm repo add kubeshark https://helm.kubeshark.co
‍helm install kubeshark kubeshark/kubeshark
```

Clean up:
```shell
helm uninstall kubeshark
```

## Building From Source

Clone this repository and run `make` command to build it. After the build is complete, the executable can be found at `./bin/kubeshark__`.

## Documentation

To learn more, read the [documentation](https://docs.kubeshark.co).

## Additional Use-cases

### Dump All Cluster-wide Traffic into a single PCAP File

Record **all** cluster traffc an incorporate into a single PCAP file (tcpdump-style)

Run Kubeshark to start capturing traffc:
```
kubeshark tap --set headless=true
```
> You can `^C` the command. Kubeshark will continue to run.

Take a snapshot of traffic (e.g. past 5 minutes):
```
kubeshark pcapdump --time 5m
```
> Read more [here](https://docs.kubeshark.co/en/pcapdump)

## Contributing

We :heart: pull requests! See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution guide.

## Code of Conduct

This project is for everyone. We ask that our users and contributors take a few minutes to review our [Code of Conduct](CODE_OF_CONDUCT.md).
