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

Kubeshark, the API Traffic Viewer for kubernetes, provides deep visibility and monitoring of all API traffic and payloads going in, out and across containers and pods inside a Kubernetes cluster.

Think of a combination of Chrome Dev Tools, TCPDump and Wireshark, re-invented for Kubernetes.

![Simple UI](https://github.com/kubeshark/assets/raw/master/png/kubeshark-ui.png)

## Download

Kubeshark uses a ~45MB pre-compiled executable binary to communicate with the Kubernetes API. We recommend downloading the `kubeshark` CLI by using one of these options:

- Choose the right binary, download and use directly from [the latest stable release](https://github.com/kubeshark/kubeshark/releases/latest).

- Use the shell script below :point_down: to automatically download the right binary for your operating system and CPU architecture:

```shell
sh <(curl -Ls https://kubeshark.co/install)
```

- Compile it from source using `make` command then use `./bin/kubeshark__` executable.

## Run

Use the `kubeshark` CLI to capture and view streaming API traffic in real time.

```shell
kubeshark tap
```

### Troubleshooting Installation
If something doesn't work or simply to play it safe prior to installing;

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

By default, Kubeshark targets the `default` namespace.
To specify a different namespace:

```
kubeshark tap -n sock-shop
```

### Specify All Namespaces

The default strategy of Kubeshark waits for the new pods
to be created. To simply tap all existing namespaces run:

```
kubeshark tap -A
```

## Documentation

Visit our documentation website: [docs.kubeshark.co](https://docs.kubeshark.co)

The documentation resources are open-source and can be found on GitHub: [kubeshark/docs](https://github.com/kubeshark/docs)

## Contributing

We ❤️ pull requests! See [CONTRIBUTING.md](CONTRIBUTING.md) for the contribution guide.

## Code of Conduct

This project is for everyone. We ask that our users and contributors take a few minutes to review our [Code of Conduct](CODE_OF_CONDUCT.md).
