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
    Want to see Kubeshark in action right now? Visit this
    <a href="https://demo.kubeshark.co/">live demo deployment</a> of Kubeshark.
  </b>
</p>

**Kubeshark** is a network observability platform for Kubernetes, providing real-time, cluster-wide visibility into Kubernetes’ network. It enables users to inspect all internal and external cluster communications, API calls, and data in transit. Additionally, Kubeshark detects anomalies and emergent behaviors, trigger autonomous remediations, and generate deep network insights.

![Simple UI](https://github.com/kubeshark/assets/raw/master/png/kubeshark-ui.png)

Think [TCPDump](https://en.wikipedia.org/wiki/Tcpdump) and [Wireshark](https://www.wireshark.org/) reimagined for Kubernetes.

#### Service-Map w/Kubernetes Context

![Service Map with Kubernetes Context](https://github.com/kubeshark/assets/raw/master/png/kubeshark-servicemap.png)

#### Cluster-Wide PCAP Recording

![Cluster-Wide PCAP Recording](https://github.com/kubeshark/assets/raw/master/png/pcap-recording.png)

## Getting Started
Download **Kubeshark**'s binary distribution [latest release](https://github.com/kubeshark/kubeshark/releases/latest) or use one of the following methods to deploy **Kubeshark**. The [web-based dashboard](https://docs.kubeshark.co/en/ui) should open in your browser, showing a real-time view of your cluster's traffic.

### Homebrew

[Homebrew](https://brew.sh/) :beer: users can install the Kubeshark CLI with:

```shell
brew install kubeshark
kubeshark tap
```

To clean up:
```shell
kubeshark clean
```

### Helm

Add the Helm repository and install the chart:

```shell
helm repo add kubeshark https://helm.kubeshark.co
helm install kubeshark kubeshark/kubeshark
```
Follow the on-screen instructions how to connect to the dashboard.

To clean up:
```shell
helm uninstall kubeshark
```

## Building From Source

Clone this repository and run the `make` command to build it. After the build is complete, the executable can be found at `./bin/kubeshark`.

## Documentation

To learn more, read the [documentation](https://docs.kubeshark.co).

## Contributing

We :heart: pull requests! See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution guide.
