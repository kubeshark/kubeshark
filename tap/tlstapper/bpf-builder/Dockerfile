FROM ubuntu

RUN apt-get update && apt-get install -y git clang llvm wget && \
	wget https://go.dev/dl/go1.17.6.linux-amd64.tar.gz && \
	tar xf go1.17.6.linux-amd64.tar.gz -C / && \
	ln -s /go/bin/go /usr/bin/go

WORKDIR /opt
ARG MIZU_BRANCH=develop
RUN git clone -b $MIZU_BRANCH https://github.com/up9inc/mizu

WORKDIR /opt/mizu
