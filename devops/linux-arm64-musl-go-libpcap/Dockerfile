FROM dockcross/linux-arm64-musl:latest AS builder-from-amd64-to-arm64v8

# Install Go
RUN curl https://go.dev/dl/go1.17.6.linux-amd64.tar.gz -Lo ./go.linux-amd64.tar.gz \
 && curl https://go.dev/dl/go1.17.6.linux-amd64.tar.gz.asc -Lo ./go.linux-amd64.tar.gz.asc \
 && curl https://dl.google.com/dl/linux/linux_signing_key.pub  -Lo linux_signing_key.pub \
 && gpg --import linux_signing_key.pub && gpg --verify ./go.linux-amd64.tar.gz.asc ./go.linux-amd64.tar.gz \
 && rm -rf /usr/local/go && tar -C /usr/local -xzf go.linux-amd64.tar.gz
ENV PATH "$PATH:/usr/local/go/bin"

# Compile libpcap from source
RUN curl https://www.tcpdump.org/release/libpcap-1.10.1.tar.gz -Lo ./libpcap.tar.gz \
 && curl https://www.tcpdump.org/release/libpcap-1.10.1.tar.gz.sig -Lo ./libpcap.tar.gz.sig \
 && curl https://www.tcpdump.org/release/signing-key.asc -Lo ./signing-key.asc \
 && gpg --import signing-key.asc && gpg --verify libpcap.tar.gz.sig libpcap.tar.gz \
 && tar -xzf libpcap.tar.gz && mv ./libpcap-* ./libpcap
WORKDIR /work/libpcap
RUN ./configure --host=arm && make \
 && cp /work/libpcap/libpcap.a /usr/xcc/aarch64-linux-musl-cross/lib/gcc/aarch64-linux-musl/*/

