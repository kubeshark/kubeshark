![Mizu: The API Traffic Viewer for Kubernetes](../assets/mizu-logo.svg)
# Istio mutual tls (mtls) with Mizu
This document describe how Mizu tapper handles workloads that configure with mtls, making the internal traffic between services in a cluster to be encrypted.

The problem happens not only with Istio, but with every service mesh that implements mtls. However, as of now Istio is the most used one, and this is why we are focusing on it.

In order to reporoduce the problem, follow those steps:
1. Deploy a sample application to a kubernetes cluster, the sample application need to do internal service to service calls
2. SSH to one of the nodes, and run tcpdump
3. Make sure you see the internal service to service calls in a plain text
4. Deploy Istio to the cluster - make sure it attached to every POD of the sample application, and also it configure with mtls (default)
5. Run tcpdump again, make sure you don't see the internal service to service calls in a plain text 

## The connection between Istio and Envoy
In order to implement its service mesh capabilities, [Istio](https://istio.io) use an [Envoy](https://www.envoyproxy.io) sidecar in front of every POD in the cluster. Envoy then responsible for the mtls comunication, and that's why we are focusing on Envoy proxy.

In the future we might see more players in that field, then we'll have to either add support for each of them. Or, go with a unified eBPF solution.

## Network namespaces
A [linux network namespace](https://man7.org/linux/man-pages/man7/network_namespaces.7.html) is an isolation that limit the process view of the network. In the container world it used to isolate one container from another. In the Kubernetes world it used to isolate a POD from another. That's mean that two containers running on the same POD share the same network namespace. A container can reach a container in the same POD by accessing `localhost`.

An Envoy proxy configured with mtls. Trap the inbound traffic directed to the POD, decrypt it and send it via `localhost` to the real container.

## Tapping mtls traffic
In order for Mizu to be able to see the decrypted traffic, it needs to listen on the same network namespace of the target POD. Multiple threads of the same process can have different network namespaces. 

[gopacket](https://github.com/google/gopacket) uses [libpacp](https://github.com/the-tcpdump-group/libpcap), by default, for capturing the traffic. Libpacap doesn't support network namespaces, and we can't ask it to listen to traffic on a different namespace. However, we can change the network namespace of the calling thread, and then start libpcap to see the traffic on a  different namespace.

## Finding the network namespace of a running process
The network namespace of a running process can be found in `/proc/PID/ns/net` link. Once we have this link, we can ask Linux to change the network namespace of a thread to this one.

This mean that Mizu needs an access to the `/proc` (procfs) of the running node.

## Finding the network namespace of a running POD
In order for Mizu to be able to listen to mtls traffic, it needs to get the PIDs of the the running PODs, filter them according to the user filters and then start listen to their internal network namespace traffic.

There is no official way in kubernetes to get from POD to PID, on purpose, The CRI implementation doesn't force a POD to be a processes on the host, it can be a Virtual Machine as well, like [Kata containers](https://katacontainers.io)

While we can provide a solution for various CRIs like Docker, Containerd, CRI-O, its better to provide with a unified solution. In order to achieve that, Mizu scan all the processes in the host, find the `envoy` processes using their `/proc/PID/exe` link.

Once Mizu have an Envoy process, it need to check whether this specific envoy process is relevant according the user filters. The user filters are a list of CLUSTER_IPS, the tapper get them via the `TapOpts.FilterAuthorities` list.

Istio send an `INSTANCE_IP` environment variable to every `envoy` proxy process. By examining the found process's environment variables we can see whether its relevant or not. Examining a process envrionment variables is done by reading the `/proc/PID/envion` file.

## Edge cases
The method we use to find envoy processes, and correlate them to the cluster ips may be wrong in certain situation. If for example a user will run an envoy process manually, and set its `INSTANCE_IP` environment variable to one of the CLUSTER_IPS the tapper get. Mizu will capute traffic for it.
