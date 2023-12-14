# PF_RING

## Overview

PF_RING™ is an advanced Linux kernel module and user-space framework designed for high-speed packet processing. It offers a uniform API for packet processing applications, enabling efficient handling of large volumes of network data.

For comprehensive information on PF_RING™, please visit the [User's Guide]((https://www.ntop.org/guides/pf_ring) and access detailed [API Documentation](http://www.ntop.org/guides/pf_ring_api/files.html).

## Provisioning mode

There are two ways to approach PF_RING kernel module load on the nodes:

1. `auto`

Kubeshark worker takes required PF_RING kernel module version from S3 and loads it on the node.

2. `kmm`

Kernel Module Management controller([KMM](https://kmm.sigs.k8s.io/documentation/deploy_kmod/)) takes required PF_RING kernel module version
from Docker container and loads it on the node.

## What mode to use?

Before deciding what methot to use, it is important to find out if there is PF_RING kernel module built already for your kernel version.
It can be done via runnig:

```
kubeshark pfring compatibility
```

This command verifies if there are kernel modules availabale for the kernel versions running on all nodes in Kubernets cluster.

### Pre-built kernel module exists and external egress allowed

If PF_RING kernel modules exist already for the target nodes, both `auto` and `kmm` provisioning modes can be used.

|auto|kmm|
|----|---|
| `SYS_MODULE` capability required for Kubeshark | `SYS_MODULE` capability required for Kubeshark|
| no additional dependencies | (!)requires `cert-manager` and `KMM` installed (follow [instructions](https://kmm.sigs.k8s.io/#installation-guide)) |
| Kubeshark falls back to `libpcap` if `PF_RING` kernel module not available | Kubshark waits until PF_RING is loaded with KMM|
| module is downloaded from S3 bucket in AWS | module is loaded from `ubehq/pf-ring-module:<kernel version>` container|
| requires egress connectivity to AWS S3 endpoints | requires egress connectivity to Docker Hub container registry|


### Pre-built kernel module doesn't exist or external egress isn't allowed

If PF_RING kernel modules don't exist yet for the target nodes, only `kmm` provisioning mode can be used (`auto` mode will still start Kubeshark, but with `libpcap`, not `PF_RING`).
This mode allows configuration of the custom container images as the source for PF_RING kernel modules.
That also means private container registries can be used to load PF_RING kernel module.

Follow these steps to use `kmm` with custom containers:

1.

Helm configuration option `tap.kernelModule.mode` controls the wa


## Appendix A: pre-build kernel versions


| Kernel version | Container |
|----------------|-----------|
|5.10.198-187.748.amzn2.x86_64|kubehq/pf-ring-module:5.10.198-187.748.amzn2.x86_64|
|5.10.199-190.747.amzn2.x86_64|kubehq/pf-ring-module:5.10.199-190.747.amzn2.x86_64|
|5.14.0-362.8.1.el9_3.x86_64|kubehq/pf-ring-module:5.14.0-362.8.1.el9_3.x86_64|
|5.15.0-1050-aws|kubehq/pf-ring-module:5.15.0-1050-aws|

## Appendix B: PF_RING kernel module custom container build

Build