![Mizu: The API Traffic Viewer for Kubernetes](../assets/mizu-logo.svg)

# Service mesh mutual tls (mtls) with Mizu

This document describe how Mizu tapper handles workloads configured with mtls, making the internal traffic between services in a cluster to be encrypted.

The list of service meshes supported by Mizu include:

- Istio
- Linkerd

## Implementation

### Istio support

#### The connection between Istio and Envoy

In order to implement its service mesh capabilities, [Istio](https://istio.io) uses an [Envoy](https://www.envoyproxy.io) sidecar in front of every pod in the cluster. The Envoy is responsible for the mtls communication, and that's why we are focusing on Envoy proxy.

In the future we might see more players in that field, then we'll have to either add support for each of them or go with a unified eBPF solution.

#### Network namespaces

A [linux network namespace](https://man7.org/linux/man-pages/man7/network_namespaces.7.html) is an isolation that limit the process view of the network. In the container world it used to isolate one container from another. In the Kubernetes world it used to isolate a pod from another. That means that two containers running on the same pod share the same network namespace. A container can reach a container in the same pod by accessing `localhost`.

An Envoy proxy configured with mtls receives the inbound traffic directed to the pod, decrypts it and sends it via `localhost` to the target container.

#### Tapping mtls traffic

In order for Mizu to be able to see the decrypted traffic it needs to listen on the same network namespace of the target pod. Multiple threads of the same process can have different network namespaces.

[gopacket](https://github.com/google/gopacket) uses [libpacp](https://github.com/the-tcpdump-group/libpcap) by default for capturing the traffic. Libpacap doesn't support network namespaces and we can't ask it to listen to traffic on a different namespace. However, we can change the network namespace of the calling thread and then start libpcap to see the traffic on a different namespace.

#### Finding the network namespace of a running process

The network namespace of a running process can be found in `/proc/PID/ns/net` link. Once we have this link, we can ask Linux to change the network namespace of a thread to this one.

This mean that Mizu needs to have access to the `/proc` (procfs) of the running node.

#### Finding the network namespace of a running pod

In order for Mizu to be able to listen to mtls traffic, it needs to get the PIDs of the the running pods, filter them according to the user filters and then start listen to their internal network namespace traffic.

There is no official way in Kubernetes to get from pod to PID. The CRI implementation purposefully doesn't force a pod to be a processes on the host. It can be a Virtual Machine as well like [Kata containers](https://katacontainers.io)

While we can provide a solution for various CRIs (like Docker, Containerd and CRI-O) it's better to have a unified solution. In order to achieve that, Mizu scans all the processes in the host, and finds the Envoy processes using their `/proc/PID/exe` link.

Once Mizu detects an Envoy process, it need to check whether this specific Envoy process is relevant according the user filters. The user filters are a list of `CLUSTER_IPS`. The tapper gets them via the `TapOpts.FilterAuthorities` list.

Istio sends an `INSTANCE_IP` environment variable to every Envoy proxy process. By examining the Envoy process's environment variables we can see whether it's relevant or not. Examining a process environment variables is done by reading the `/proc/PID/envion` file.

#### Edge cases

The method we use to find Envoy processes and correlate them to the cluster IPs may be inaccurate in certain situations. If, for example, a user runs an Envoy process manually, and set its `INSTANCE_IP` environment variable to one of the `CLUSTER_IPS` the tapper gets, then Mizu will capture traffic for it.

## Development

In order to create a service mesh setup for development, follow those steps:

1. Deploy a sample application to a Kubernetes cluster, the sample application needs to make internal service to service calls
2. SSH to one of the nodes, and run `tcpdump`
3. Make sure you see the internal service to service calls in a plain text
4. Deploy a service mesh (Istio, Linkerd) to the cluster - make sure it is attached to all pods of the sample application, and that it is configured with mtls (default)
5. Run `tcpdump` again, make sure you don't see the internal service to service calls in a plain text
