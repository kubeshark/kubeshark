FROM messense/rust-musl-cross:x86_64-musl AS builder-from-arm64v8-to-amd64

ENV CROSS_TRIPLE x86_64-unknown-linux-musl
ENV CROSS_ROOT /usr/local/musl

ENV AS=${CROSS_ROOT}/bin/${CROSS_TRIPLE}-as \
    AR=${CROSS_ROOT}/bin/${CROSS_TRIPLE}-ar \
    CC=${CROSS_ROOT}/bin/${CROSS_TRIPLE}-gcc \
    CPP=${CROSS_ROOT}/bin/${CROSS_TRIPLE}-cpp \
    CXX=${CROSS_ROOT}/bin/${CROSS_TRIPLE}-g++ \
    LD=${CROSS_ROOT}/bin/${CROSS_TRIPLE}-ld \
    FC=${CROSS_ROOT}/bin/${CROSS_TRIPLE}-gfortran

# Install Go
WORKDIR /
RUN curl https://go.dev/dl/go1.17.6.linux-arm64.tar.gz -Lo ./go.linux-arm64.tar.gz \
 && curl https://go.dev/dl/go1.17.6.linux-arm64.tar.gz.asc -Lo ./go.linux-arm64.tar.gz.asc \
 && curl https://dl.google.com/dl/linux/linux_signing_key.pub  -Lo linux_signing_key.pub \
 && gpg --import linux_signing_key.pub && gpg --verify ./go.linux-arm64.tar.gz.asc ./go.linux-arm64.tar.gz \
 && rm -rf /usr/local/go && tar -C /usr/local -xzf go.linux-arm64.tar.gz
ENV PATH "$PATH:/usr/local/go/bin"

# Compile libpcap from source
RUN curl https://www.tcpdump.org/release/libpcap-1.10.1.tar.gz -Lo ./libpcap.tar.gz \
 && curl https://www.tcpdump.org/release/libpcap-1.10.1.tar.gz.sig -Lo ./libpcap.tar.gz.sig \
 && curl https://www.tcpdump.org/release/signing-key.asc -Lo ./signing-key.asc \
 && gpg --import signing-key.asc && gpg --verify libpcap.tar.gz.sig libpcap.tar.gz \
 && tar -xzf libpcap.tar.gz && mv ./libpcap-* ./libpcap
WORKDIR /libpcap
RUN ./configure --host=x86_64 && make \
 && cp /libpcap/libpcap.a /usr/local/musl/lib/gcc/x86_64-unknown-linux-musl/*/
