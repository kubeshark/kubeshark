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
    <a href="https://join.slack.com/t/kubeshark/shared_invite/zt-3jdcdgxdv-1qNkhBh9c6CFoE7bSPkpBQ">
      <img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-green?logo=Slack&style=flat-square&label=slack">
    </a>
</p>
<p align="center">
  <b>
    Want to see Kubeshark in action right now? Visit this
    <a href="https://demo.kubeshark.com/">live demo deployment</a> of Kubeshark.
  </b>
</p>

> Latest release enables free deployment on clusters with up to 100 nodes.

**Kubeshark** is an API traffic analyzer for Kubernetes, providing deep packet inspection with complete API and Kubernetes contexts, retaining cluster-wide L4 traffic (PCAP), and using minimal production compute resources.

![Simple UI](https://github.com/kubeshark/assets/raw/master/png/kubeshark-ui.png)

Think [TCPDump](https://en.wikipedia.org/wiki/Tcpdump) and [Wireshark](https://www.wireshark.org/) reimagined for Kubernetes.

Access cluster-wide PCAP traffic by pressing a single button, without the need to install `tcpdump` or manually copy files. Understand the traffic context in relation to the API and Kubernetes contexts.

#### Service-Map w/Kubernetes Context

![Service Map with Kubernetes Context](https://github.com/kubeshark/assets/raw/master/png/kubeshark-servicemap.png)

#### Export Cluster-Wide L4 Traffic (PCAP)

Imagine having a cluster-wide [TCPDump](https://www.tcpdump.org/)-like capability—exporting a single [PCAP](https://www.ietf.org/archive/id/draft-gharris-opsawg-pcap-01.html) file that consolidates traffic from multiple nodes, all accessible with a single click.

1. Go to the **Snapshots** tab
2. Create a new snapshot
3. **Optionally** select the nodes (default: all nodes)
4. **Optionally** select the time frame (default: last one hour)
5. Press **Create**

<img width="3342" height="1206" alt="image" src="https://github.com/user-attachments/assets/e8e47996-52b7-4028-9698-f059a13ffdb7" />


Once the snapshot is ready, click the PCAP file to export its contents and open it in Wireshark.

## Getting Started
Download **Kubeshark**'s binary distribution [latest release](https://github.com/kubeshark/kubeshark/releases/latest) or use one of the following methods to deploy **Kubeshark**. The [web-based dashboard](https://docs.kubeshark.com/en/ui) should open in your browser, showing a real-time view of your cluster's traffic.

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
helm repo add kubeshark https://helm.kubeshark.com
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

To learn more, read the [documentation](https://docs.kubeshark.com).

## Contributing

We :heart: pull requests! See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution guide.
