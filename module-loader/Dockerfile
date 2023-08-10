### Builder image
ARG TARGETARCH=amd64
FROM ${TARGETARCH}/debian:bullseye AS builder

ARG KERNEL_VERSION

RUN apt-get update
RUN apt-get install -y -q \
    curl \
    build-essential \
    linux-headers-${KERNEL_VERSION}

WORKDIR /work
RUN curl https://github.com/ntop/PF_RING/archive/refs/tags/8.4.0.tar.gz -Lo ./pfring.tar.xz && \
    tar -xf pfring.tar.xz && \
    mv ./PF_RING-* ./pfring

WORKDIR /work/pfring/kernel
ENV KERNELRELEASE=${KERNEL_VERSION}
RUN make


### The shipped image
ARG TARGETARCH=amd64
FROM ${TARGETARCH}/debian:bullseye

ARG KERNEL_VERSION

RUN apt-get update && apt-get install -y kmod

COPY --from=builder /work/pfring/kernel/pf_ring.ko /opt/lib/modules/${KERNEL_VERSION}/

RUN depmod -b /opt ${KERNEL_VERSION}
