# Copyright (c) 2016-2017 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

FROM ubuntu

ENV GOPATH /opt/go
ENV GOROOT /usr/local/go
ENV PATH $PATH:$GOROOT/bin/:$GOPATH/bin/

RUN apt-get update \
    && apt-get install -yqq \
       build-essential \
       wget \
       tar \
       git \
    && rm -rf /var/lib/apt/lists/* \
    && wget https://storage.googleapis.com/golang/go1.8.3.linux-amd64.tar.gz \
    && tar -C /usr/local/ -xzvf go1.8.3.linux-amd64.tar.gz \
    && mkdir -p /opt/go/src/ /opt/go/pkg/ /opt/go/bin/

RUN mkdir -p /dev/net \
    && mknod /dev/net/tun c 10 200 \
    && chmod 0666 /dev/net/tun
