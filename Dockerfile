FROM ubuntu

RUN apt-get update \
    && apt-get install -yqq \
        mtr \
        tcpdump \
        iperf3 \
        iproute2 \
        iputils-ping \
    && rm -rf /var/lib/apt/lists/*

COPY ./bin/start_quantum.sh /bin/start_quantum.sh

RUN chmod 770 /bin/start_quantum.sh
