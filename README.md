<p align="center">
  <img src="assets/kubeshark-logo.svg" alt="Kubeshark: Kubernetes deep visibility." height="128px"/>
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
    <a href="https://join.slack.com/t/up9/shared_invite/zt-tfjnduli-QzlR8VV4Z1w3YnPIAJfhlQ">
      <img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social">
    </a>
</p>

Kubeshark is an **observability and monitoring tool for** [**Kubernetes**](https://kubernetes.io/), enabling **dynamic analysis** of your microservices, detecting **anomalies** and **triggering functions** when certain patterns appear in runtime.

Think of Kubeshark as a **Kubernetes-aware** combination of [**Wireshark**](https://www.wireshark.org/), [**BPF Compiler Collection (BCC) tools**](https://github.com/iovisor/bcc) and beyond.

![Simple UI](assets/kubeshark-ui.png)

## Quickstart

### Install

Install Kubeshark CLI with your favorite package manager:

#### Homebrew

```shell
brew install kubeshark
```

#### apt

```shell
apt install kubeshark
```

#### yum

```shell
yum install kubeshark
```

#### apk

```shell
apk add kubeshark
```

#### Snap

```shell
snap install kubeshark
```

#### AppImage

```shell
appimage install kubeshark
```

#### Flatpak

```shell
flatpak install kubeshark
```

#### Chocolatey

```shell
choco install kubeshark
```

### Deploy

Once you have the Kubeshark CLI installed on your system.
Run the command below to deploy Kubeshark Agent into your Kubernetes cluster.

> Kubeshark images are hosted on Docker Hub. Make sure you have access to https://hub.docker.com/

> Make sure `kubeshark` executable in your `PATH`.

```shell
kubeshark deploy
```

## Documentation

Visit our documentation website: [docs.kubeshark.co](https://docs.kubeshark.co)

The documentation resources are open-source and can be found on GitHub: [kubeshark/docs](https://github.com/kubeshark/docs)

## Contributing

We ❤️ pull requests! See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for the contribution guide.

## Code of Conduct

This project is for everyone. We ask that our users and contributors take a few minutes to review our [Code of Conduct](docs/CODE_OF_CONDUCT.md).
