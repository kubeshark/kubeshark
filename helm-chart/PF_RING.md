# PF_RING

<!-- TOC -->

- [PF\_RING](#pf_ring)
  - [Overview](#overview)
  - [Loading PF\_RING module on Kubernetes nodes](#loading-pf_ring-module-on-kubernetes-nodes)
    - [Pre-built kernel module exists and external egress allowed](#pre-built-kernel-module-exists-and-external-egress-allowed)
    - [Pre-built kernel module doesn't exist or external egress isn't allowed](#pre-built-kernel-module-doesnt-exist-or-external-egress-isnt-allowed)
  - [Appendix A: PF\_RING kernel module compilation](#appendix-a-pf_ring-kernel-module-compilation)
    - [Automated complilation](#automated-complilation)
    - [Manual compilation](#manual-compilation)

<!-- /TOC -->

## Overview

PF_RING™ is an advanced Linux kernel module and user-space framework designed for high-speed packet processing. It offers a uniform API for packet processing applications, enabling efficient handling of large volumes of network data.

For comprehensive information on PF_RING™, please visit the [User's Guide]((https://www.ntop.org/guides/pf_ring) and access detailed [API Documentation](http://www.ntop.org/guides/pf_ring_api/files.html).

## Loading PF_RING module on Kubernetes nodes

PF_RING kernel module loading is performed via of the `worker` component pod.
The target container `tap.kernelModule.image` must contain `pf_ring.ko` file under path `/opt/lib/modules/<kernel version>/pf_ring.ko`.
Kubeshark provides ready to use containers with kernel modules for the most popular kernel versions running in different managed clouds.

Prior to deploying `kubeshark` with PF_RING enabled, it is essential to verify if a PF_RING kernel module is already built for your kernel version.
Kubeshark provides additional CLI tool for this purpose - [pf-ring-compiler](https://github.com/kubeshark/pf-ring-compiler).

Compatibility verification can be done by running:

```bash
pfring-compiler compatibility
```

This command checks for the availability of kernel modules for the kernel versions running across all nodes in the Kubernetes cluster.

Example output for a compatible cluster:

```bash
Node                                          Kernel Version                 Supported
ip-192-168-77-230.us-west-2.compute.internal  5.10.199-190.747.amzn2.x86_64  true
ip-192-168-34-216.us-west-2.compute.internal  5.10.199-190.747.amzn2.x86_64  true

Cluster is compatible
```

Another option to verify availability of kernel modules is just inspecting available kernel module versions via:

```bash
curl https://api.kubeshark.co/kernel-modules/meta/versions.jso
```

Based on Kubernetes cluster compatibility and external connection capabilities, user has two options:

1. Use Kubeshark provided container `kubeshark/pf-ring-module`
2. Build custom container with required kernel module version. 

### Pre-built kernel module exists and external egress allowed

In this case no additional configuration required.
Kubeshark will load PF_RING kernel module from the default `kubeshark/pf-ring-module:all` container.

### Pre-built kernel module doesn't exist or external egress isn't allowed

In this case building custom Docker image is required.

1. Compile PF_RING kernel module for target version

Skip if you have `pf_ring.ko` for the target kernel version.
Otherwise, follow [Appendix A](#appendix-a-pf_ring-kernel-module-compilation) for details.

2. Build container

The same build process Kubeshark has can be reused (follow [pfring-compilier](https://github.com/kubeshark/pf-ring-compiler/tree/main/modules) for details).

3. Configure Helm values

```yaml
tap:
  kernelModule:
    image: <container from stage 2>
```


## Appendix A: PF_RING kernel module compilation

PF_RING kernel module compilation can be completed automatically or manually.

### Automated complilation

In case your Kubernetes workers run supported Linux distribution, `kubeshark` CLI can be used to build PF_RING module:

```bash
pfring-compiler compile --target <distro>
```

This command requires:

- kubectl to be installed and configured with a proper context
- egress connection to Internet available

This command:

1. Runs Kubernetes job with build container
2. Waits for job to be completed
3. Downloads `pf-ring-<kernel version>.ko` file into the current folder.
4. Cleans up created job.

Currently supported distros:

- Ubuntu
- RHEL 9
- Amazon Linux 2

### Manual compilation

The process description is based on Ubuntu 22.04 distribution.

1. Get terminal access to the node with target kernel version
This can be done either via SSH directly to node or with debug container running on the target node:

```bash
kubectl debug node/<target node> -it --attach=true --image=ubuntu:22.04
```

2. Install build tools and kernel headers

```bash
apt update
apt install -y gcc build-essential make git wget tar gzip
apt install -y linux-headers-$(uname -r)
```

3. Download PF_RING source code

```bash
wget https://github.com/ntop/PF_RING/archive/refs/tags/8.4.0.tar.gz
tar -xf 8.4.0.tar.gz
cd PF_RING-8.4.0/kernel
```

4. Compile the kernel module

```bash
make KERNEL_SRC=/usr/src/linux-headers-$(uname -r)
```

5. Copy `pf_ring.ko` to the local file system.

Use `scp` or `kubectl cp` depending on type of access(SSH or debug pod).
