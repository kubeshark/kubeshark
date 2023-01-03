<p align="center">
  <img src="https://raw.githubusercontent.com/kubeshark/assets/master/svg/kubeshark-logo.svg" alt="Kubeshark: Traffic viewer for Kubernetes." height="128px"/>
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
		<a href="https://discord.gg/WkvRGMUcx7">
      <img alt="Discord" src="https://img.shields.io/discord/1042559155224973352?logo=Discord&style=flat-square&label=discord">
    </a>
    <a href="https://join.slack.com/t/kubeshark/shared_invite/zt-1m90td3n7-VHxN_~V5kVp80SfQW3SfpA">
      <img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-green?logo=Slack&style=flat-square&label=slack">
    </a>
</p>

<p>
<p align="center"><b>
<a href="https://github.com/kubeshark/kubeshark/releases/latest">V38</a> is out with <a href="https://docs.kubeshark.co/en/pcap">PCAP</a>, <a href="https://docs.kubeshark.co/en/tcp">TCP streams</a>, <a href="https://docs.kubeshark.co/en/history">Historic Traffic Snapshot</a> and so much more. Read about it <a href="https://kubeshark.co/pcap-or-it-didnt-happen">here</a>.
	</b></p>

**Kubeshark** is an API Traffic Viewer for [Kubernetes](https://kubernetes.io/) providing deep visibility and monitoring of all API traffic and payloads going in, out and across containers and Pods inside a Kubernetes cluster.

![Simple UI](https://github.com/kubeshark/assets/raw/master/png/kubeshark-ui.png)

Think [TCPDump](https://en.wikipedia.org/wiki/Tcpdump) and [Wireshark](https://www.wireshark.org/) re-invented for Kubernetes

## Getting Started

Download **Kubeshark**'s binary distribution [latest release](https://github.com/kubeshark/kubeshark/releases/latest) and run following one of these examples:

```shell
kubeshark tap
```
```
kubeshark tap -A
```
```
kubeshark tap -n sock-shop "(catalo*|front-end*)"
```
Running any of the :point_up: above commands will open a local [Web UI](https://docs.kubeshark.co/en/ui) immediately showing Kubernetes trafic streaming in real time.

## Homebrew
MacOS and GNU/Linux users available way to install via [Homebrew](https://brew.sh/):

```bash
# Tap a new formula:
brew tap kubeshark/kubeshark

# Installation:
brew install kubeshark
```

## Documentation

To learn more, read the [documentation](https://docs.kubeshark.co).

## Contributing

We ❤️ pull requests! See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution guide.

## Code of Conduct

This project is for everyone. We ask that our users and contributors take a few minutes to review our [Code of Conduct](CODE_OF_CONDUCT.md).
