FROM ubuntu

RUN apt-get update \
    && apt-get install -yqq \
        tcpdump \
        iperf3 \
        iproute2 \
        iputils-ping \
    && rm -rf /var/lib/apt/lists/*
