# Copyright (c) 2016-2018 Christian Saide <supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

FROM ubuntu

RUN apt-get update \
    && apt-get install -yqq \
        mtr \
        tcpdump \
        iperf3 \
        iproute2 \
        iputils-ping \
        net-tools \
        hping3 \
        iptables \
        dnsutils \
        curl \
        wget \
    && rm -rf /var/lib/apt/lists/*

COPY ./start_quantum.sh /bin/start_quantum.sh

RUN chmod 770 /bin/start_quantum.sh
