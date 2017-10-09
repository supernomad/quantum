##########
 Security
##########

Security within ``quantum`` is fully configurable and ranges from plain text communication, to fully encrypted and authenticated communication between peers participating in the network. In order to fully secure the network please carefully read and understand this document.

Etcd
====

``quantum`` requires Etcd to distribute networking mappings among the peers participating within the greater network. While there is never any private security related information stored within the cluster, the entire network map is stored along with public IP's and public ports of each peer.

There are a few things to keep in mind when handling the Etcd cluster that ``quantum`` is going to utilize. In order to provide the most security the cluster should be configured with `TLS enabled <https://coreos.com/etcd/docs/latest/op-guide/security.html>`_ for both client and peer communication with certificate verification enabled. The TLS certificates should ideally be unique for each client and server, and be signed by a strong CA. This is critical for ``quantum`` to guarantee safe operation over the WAN.

Etcd `authentication <https://coreos.com/etcd/docs/latest/op-guide/authentication.html>`_ should be enabled as well, but is not required given TLS certificates are unique within your cluster and verification is enabled.

Network
=======

To provide network level security ``quantum`` exposes two different methodologies.

  #. DTLS
  #. Packet Encryption

Each methodology comes with its own costs and benefits. They also operate at different levels within ``quantum``, which allow for varying levels of granularity and configuration requirements. Which one to use will come down to the security needs of the infrastructure that ``quantum`` will deployed on.

DTLS
----

The DTLS module is a networking backend and utilizes OpenSSL to both authenticate and encrypt data between peers with perfect forward secrecy. While this module provides the most security for ``quantum``, it also comes with the costs of applying to **all** peers in the ``quantum`` network, and requiring more initial setup/configuration.

The configuration options for DTLS should be carefully reviewed before enabling this functionality to ensure that it operates correctly. In general the DTLS module should be supplied with unique strongly signed certificates and verification should **always** be enabled.

Packet Encryption
-----------------

The packet encryption module is a plugin that utilizes a combination of `pbkdf2 <https://en.wikipedia.org/wiki/PBKDF2>`_, `curve25519 <https://en.wikipedia.org/wiki/Curve25519>`_, and `AES256-GCM <https://en.wikipedia.org/wiki/Galois/Counter_Mode>`_, in order to provide authenticated and encrypted peer communication. This module is less secure than the DTLS module, but comes with the benfits of applying to only specific peers, and having no additional setup. The difference in security between this module and the DTLS module is that this module does not have perfect forward secrecy. The shared security, while unique for each pair of commuincating peers, is generated at startup and will persist until one of the communicating peers is restarted.

There is no configuration required to utilize the packet encryption module, other than enabling the plugin on the desired peers.
