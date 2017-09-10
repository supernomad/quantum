# Copyright (c) 2016-2017 Christian Saide <supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

FROM ubuntu

ENV GOVERSION 1.9
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
    && wget https://storage.googleapis.com/golang/go${GOVERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local/ -xzvf go${GOVERSION}.linux-amd64.tar.gz \
    && mkdir -p /opt/go/src/ /opt/go/pkg/ /opt/go/bin/
