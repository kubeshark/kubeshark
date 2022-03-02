![Mizu: The API Traffic Viewer for Kubernetes](assets/mizu-logo.svg)

<p align="center">
    <a href="https://github.com/up9inc/mizu/blob/main/LICENSE">
        <img alt="GitHub License" src="https://img.shields.io/github/license/up9inc/mizu?logo=GitHub&style=flat-square">
    </a>
    <a href="https://github.com/up9inc/mizu/releases/latest">
        <img alt="GitHub Latest Release" src="https://img.shields.io/github/v/release/up9inc/mizu?logo=GitHub&style=flat-square">
    </a>
    <a href="https://hub.docker.com/r/up9inc/mizu">
      <img alt="Docker pulls" src="https://img.shields.io/docker/pulls/up9inc/mizu?color=%23099cec&logo=Docker&style=flat-square">
    </a>
    <a href="https://hub.docker.com/r/up9inc/mizu">
      <img alt="Image size" src="https://img.shields.io/docker/image-size/up9inc/mizu/latest?logo=Docker&style=flat-square">
    </a>
    <a href="https://join.slack.com/t/up9/shared_invite/zt-tfjnduli-QzlR8VV4Z1w3YnPIAJfhlQ">
      <img alt="Slack" src="https://img.shields.io/badge/slack-join_chat-white.svg?logo=slack&style=social">
    </a>
</p>

# The API Traffic Viewer for Kubernetes

A simple-yet-powerful API traffic viewer for Kubernetes enabling you to view all API communication between microservices to help your debug and troubleshoot regressions.

Think TCPDump and Wireshark re-invented for Kubernetes.

![Simple UI](assets/mizu-ui.png)

## Quickstart and documentation

You can run Mizu on any Kubernetes cluster (version of 1.16.0 or higher) in a matter of seconds. See the [Mizu Getting Started Guide](https://getmizu.io/docs/) for how.

For more comprehensive documentation, start with the [docs](https://getmizu.io/docs/mizu/mizu-cli).

## Working in this repo

We ❤️ pull requests! See [CONTRIBUTING.md](docs/CONTRIBUTING.md) for info on contributing changes.



## Code of Conduct

This project is for everyone. We ask that our users and contributors take a few minutes to review our [Code of Conduct](docs/CODE_OF_CONDUCT.md).









## Features

- Simple and powerful CLI
- Monitoring network traffic in real-time. Supported protocols:
  - [HTTP/1.x](https://datatracker.ietf.org/doc/html/rfc2616) (REST, GraphQL, SOAP, etc.)
  - [HTTP/2](https://datatracker.ietf.org/doc/html/rfc7540) (gRPC)
  - [AMQP](https://www.rabbitmq.com/amqp-0-9-1-reference.html) (RabbitMQ, Apache Qpid, etc.)
  - [Apache Kafka](https://kafka.apache.org/protocol)
  - [Redis](https://redis.io/topics/protocol)
- Works with Kubernetes APIs. No installation or code instrumentation
- Rich filtering

## Requirements

A Kubernetes server version of 1.16.0 or higher is required.

## Download

Download Mizu for your platform and operating system

### Latest Stable Release

* for MacOS - Intel 
```
curl -Lo mizu \
https://github.com/up9inc/mizu/releases/latest/download/mizu_darwin_amd64 \
&& chmod 755 mizu
```
 
* for Linux - Intel 64bit
```
curl -Lo mizu \
https://github.com/up9inc/mizu/releases/latest/download/mizu_linux_amd64 \
&& chmod 755 mizu
``` 

SHA256 checksums are available on the [Releases](https://github.com/up9inc/mizu/releases) page

### Development (unstable) Build
Pick one from the [Releases](https://github.com/up9inc/mizu/releases) page

## How to Run

1. Find pods you'd like to tap to in your Kubernetes cluster
2. Run `mizu tap` or `mizu tap PODNAME`
3. Open browser on `http://localhost:8899` **or** as instructed in the CLI
4. Watch the API traffic flowing
5. Type ^C to stop

## Examples

Run `mizu help` for usage options

To tap all pods in current namespace - 
``` 
 $ kubectl get pods 
 NAME                            READY   STATUS    RESTARTS   AGE
 carts-66c77f5fbb-fq65r          2/2     Running   0          20m
 catalogue-5f4cb7cf5-7zrmn       2/2     Running   0          20m
 front-end-649fc5fd6-kqbtn       2/2     Running   0          20m
 ..

 $ mizu tap
 +carts-66c77f5fbb-fq65r
 +catalogue-5f4cb7cf5-7zrmn
 +front-end-649fc5fd6-kqbtn
 Web interface is now available at http://localhost:8899
 ^C
```


### To tap specific pod
```bash
 $ kubectl get pods 
 NAME                            READY   STATUS    RESTARTS   AGE
 front-end-649fc5fd6-kqbtn       2/2     Running   0          7m
 ..

 $ mizu tap front-end-649fc5fd6-kqbtn
 +front-end-649fc5fd6-kqbtn
 Web interface is now available at http://localhost:8899
 ^C
```

### To tap multiple pods using regex
```bash
 $ kubectl get pods 
 NAME                            READY   STATUS    RESTARTS   AGE
 carts-66c77f5fbb-fq65r          2/2     Running   0          20m
 catalogue-5f4cb7cf5-7zrmn       2/2     Running   0          20m
 front-end-649fc5fd6-kqbtn       2/2     Running   0          20m
 ..

 $ mizu tap "^ca.*"
 +carts-66c77f5fbb-fq65r
 +catalogue-5f4cb7cf5-7zrmn
 Web interface is now available at http://localhost:8899
 ^C
```

## Configuration

Mizu can optionally work with a config file that can be provided as a CLI argument (using `--set config-path=<PATH>`) or if not provided, will be stored at ${HOME}/.mizu/config.yaml 
In case of partial configuration defined, all other fields will be used with defaults <br />
You can always override the defaults or config file with CLI flags

To get the default config params run `mizu config` <br />
To generate a new config file with default values use `mizu config -r`


## Advanced Usage

### Kubeconfig

It is possible to change the kubeconfig path using `KUBECONFIG` environment variable or the command like flag
with `--set kube-config-path=<PATH>`. </br >
If both are not set - Mizu assumes that configuration is at `${HOME}/.kube/config`

### Namespace-Restricted Mode

Some users have permission to only manage resources in one particular namespace assigned to them
By default `mizu tap` creates a new namespace `mizu` for all of its Kubernetes resources. In order to instead install
Mizu in an existing namespace, set the `mizu-resources-namespace` config option

If `mizu-resources-namespace` is set to a value other than the default `mizu`, Mizu will operate in a
Namespace-Restricted mode. It will only tap pods in `mizu-resources-namespace`. This way Mizu only requires permissions
to the namespace set by `mizu-resources-namespace`. The user must set the tapped namespace to the same namespace by
using the `--namespace` flag or by setting `tap.namespaces` in the config file

Setting `mizu-resources-namespace=mizu` resets Mizu to its default behavior

For detailed list of k8s permissions see [PERMISSIONS](docs/PERMISSIONS.md) document

### User agent filtering

User-agent filtering (like health checks) - can be configured using command-line options:

```shell
$ mizu tap "^ca.*" --set tap.ignored-user-agents=kube-probe --set tap.ignored-user-agents=prometheus
+carts-66c77f5fbb-fq65r
+catalogue-5f4cb7cf5-7zrmn
Web interface is now available at http://localhost:8899
^C

```
Any request that contains `User-Agent` header with one of the specified values (`kube-probe` or `prometheus`) will not be captured

