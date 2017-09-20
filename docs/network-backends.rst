##################
 Network Backends
##################

There are multiple networking backends that ``quantum`` implements which offer different levels of security and ease of operation. That being said under the hood each of the network backends end up transmitting UDP packets in one form or another.

DTLS
====

Backend Name:
  dtls

The DTLS network backend is an implementation of DTLS utilizing OpenSSL. This backend requires extra configuration for each peer in the network in the form of TLS certificates and enabling/disabling certification verfication. When utilizing this backend all network traffic is encrypted and authenticated before being transmitted as UDP packets using standardized functionality within DTLS and is fully compliant with their specifications.

  Note: Always enable certificate verficiation to ensure security within ``quantum``.

One caveat to bear in mind when utlizing the DTLS backend, is that OpenSSL is statically compiled into ``quantum``. This has two rammifcations, first it means that the latest version of ``quantum`` tracks with the latest version of OpenSSL, second to update OpenSSL ``quantum`` itself must be updated.

UDP
===

Backend Name:
  udp

The UDP network backend is a simple socket implementation that sends plain text packets to other peers in the network. This can and should be combined with the packet encryption plugin to ensure that data is properly secured in transit.
