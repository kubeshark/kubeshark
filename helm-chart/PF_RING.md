# PF_RING

<!-- TOC -->

- [PF_RING](#pf_ring)
    - [Overview](#overview)
    - [Provisioning mode](#provisioning-mode)
    - [Selection of Provisioning Mode](#selection-of-provisioning-mode)
        - [Pre-built kernel module exists and external egress allowed](#pre-built-kernel-module-exists-and-external-egress-allowed)
        - [Pre-built kernel module doesn't exist or external egress isn't allowed](#pre-built-kernel-module-doesnt-exist-or-external-egress-isnt-allowed)
            - [Steps to Use kmm with Custom Containers](#steps-to-use-kmm-with-custom-containers)
    - [Appendix A: PF_RING kernel module compilation](#appendix-a-pf_ring-kernel-module-compilation)
        - [Automated complilation](#automated-complilation)
        - [Manual compilation](#manual-compilation)

<!-- /TOC -->

## Overview

PF_RING™ is an advanced Linux kernel module and user-space framework designed for high-speed packet processing. It offers a uniform API for packet processing applications, enabling efficient handling of large volumes of network data.

For comprehensive information on PF_RING™, please visit the [User's Guide]((https://www.ntop.org/guides/pf_ring) and access detailed [API Documentation](http://www.ntop.org/guides/pf_ring_api/files.html).

## Provisioning mode

There are two approaches for loading the PF_RING kernel module on nodes:

1. `auto`

In this mode, the Kubeshark worker retrieves the necessary PF_RING kernel module version from an S3 bucket and loads it onto the node.

> mode=auto requires an active internet connection and is not suitable for air-gapped environments.

2. `kmm`

The Kernel Module Management controller ([KMM](https://kmm.sigs.k8s.io/documentation/deploy_kmod/)) acquires the required PF_RING kernel module version from a Docker container and loads it onto the node

> mode=kmm is suitable for air-gapped environments.

## Selection of Provisioning Mode

> This step is optional. In case mode=auto and no PF_RING kernel module is found Kubeshark falls back to `libpcap` if `PF_RING` kernel module not available

Prior to choosing a method, it is essential to verify if a PF_RING kernel module is already built for your kernel version.
Kubeshark provides additional CLI tool for this purpose - [pf-ring-compiler](https://github.com/kubeshark/pf-ring-compiler).

Compatibility verification can be done by running:

```
pfring-compiler compatibility
```

This command checks for the availability of kernel modules for the kernel versions running across all nodes in the Kubernetes cluster.

Example output for a compatible cluster:
```
Node                                          Kernel Version                 Supported
ip-192-168-77-230.us-west-2.compute.internal  5.10.199-190.747.amzn2.x86_64  true
ip-192-168-34-216.us-west-2.compute.internal  5.10.199-190.747.amzn2.x86_64  true

Cluster is compatible
```


### Pre-built kernel module exists and external egress allowed

If PF_RING kernel modules are already available for the target nodes (cluster is compatible), both `auto` and `kmm` modes are applicable.

|auto|kmm|
|----|---|
| `SYS_MODULE` capability required for Kubeshark | `SYS_MODULE` capability is **not** required for Kubeshark|
| no additional dependencies | (!)requires `cert-manager` and `KMM` installed (follow [instructions](https://kmm.sigs.k8s.io/documentation/install/) or a specific [cloud platform guide](https://kmm.sigs.k8s.io/lab/)) |
| Kubeshark falls back to `libpcap` if `PF_RING` kernel module not available | Kubshark waits until PF_RING is loaded with KMM|
| module is downloaded from S3 bucket in AWS | module is loaded from `kubeshark/pf-ring-module:<kernel version>` container|
| requires egress connectivity to AWS S3 endpoints | can work in an air-gapped environment when the docker images are stored in a local container registry|


### Pre-built kernel module doesn't exist or external egress isn't allowed

In cases where PF_RING kernel modules are not yet available for the target nodes, or if external egress is restricted, the `kmm` mode is the only viable option (`auto` mode would start Kubeshark with libpcap, not PF_RING).
This approach enables the use of custom container images as the source for PF_RING kernel modules and allows leveraging private container registries.

#### Steps to Use kmm with Custom Containers

1. Compile the pf_ring.ko kernel module for your target kernel version (see [Appendix B](#appendix-b-pf_ring-kernel-module-compilation) for instructions).

After building the module with kubeshark pfring compile, you will obtain a `pf-ring-<kernel version>.ko` file.
If manually built, rename the kernel module to this format.

2. Build and push Docker container(-s) with the kernel module file from stage 1.

Create `Dockerfile` in the folder with PF_RING kernel module:

```
FROM alpine:3.18
ARG KERNEL_VERSION

COPY pf-ring-${KERNEL_VERSION}.ko /opt/lib/modules/${KERNEL_VERSION}/pf_ring.ko
RUN apk add kmod

RUN depmod -b /opt ${KERNEL_VERSION}
```

Run build&command:

```
docker build --build-arg <kernel version> <your registry>/<image>:<kernel version>
docker push <your registry>:/<image>:<kernel version>
```

It is recommended to use kernel version as a container tag for consistency.


3. Configure Helm values

```
tap:
  kernelModule:
    mode: kmm
    kernelMappings:
    - regexp: '<kernel version>'
      containerImage: '<your-registry>/<image>:<kernel version>'
    imageRepoSecret: <optional secret with credentials for private registry>
```


## Appendix A: PF_RING kernel module compilation

PF_RING kernel module compilation can be completed automatically or manually.

### Automated complilation

In case your Kubernetes workers run supported Linux distribution, `kubeshark` CLI can be used to build PF_RING module:

```
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

```
kubectl debug node/<target node> -it --attach=true --image=ubuntu:22.04
```

2. Install build tools and kernel headers

```
apt update
apt install -y gcc build-essential make git wget tar gzip
apt install -y linux-headers-$(uname -r)
```

3. Download PF_RING source code

```
wget https://github.com/ntop/PF_RING/archive/refs/tags/8.4.0.tar.gz
tar -xf 8.4.0.tar.gz
cd PF_RING-8.4.0/kernel
```

4. Compile the kernel module

```
make KERNEL_SRC=/usr/src/linux-headers-$(uname -r)
```

5. Copy `pf_ring.ko` to the local file system.

Use `scp` or `kubectl cp` depending on type of access(SSH or debug pod).
