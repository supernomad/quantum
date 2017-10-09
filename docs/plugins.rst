#########
 Plugins
#########

There are various plugins that can be enabled within ``quantum`` to augment each packet traversing the network. These plugins can be enabled on a per peer basis, but it should be kept in mind that a plugin is only utilized when the two communicating peers **both** have the plugin enabled.

This means that if nodeA and nodeB have the ``compression`` plugin enabled but nodeC does not, only traffic between nodeA and nodeB will be compressed will communication of any kind to nodeC will be uncompressed.

Compression
===========

Plugin Name:
  compression

The compression plugin uses the `snappy <https://google.github.io/snappy/>`_ compression algorithm to encode and decode packets as they are sent and received. Depending on the types of traffic being routed over ``quantum`` this can result in a massive bandwidth savings, at the cost of CPU cycles to compute the compression. However given the data is incompressible, it will be skipped over and little if any CPU will be consumed.

Packet Encryption
=================

Plugin Name:
  encryption

The encryption plugin utilizes a combination of `pbkdf2 <https://en.wikipedia.org/wiki/PBKDF2>`_, `curve25519 <https://en.wikipedia.org/wiki/Curve25519>`_, and `AES256-GCM <https://en.wikipedia.org/wiki/Galois/Counter_Mode>`_ to encrypt and authenticate packets aas they are sent and received. This plugin is an alternative to using the DTLS networking backend, and will allow for granular application depending on which peers in the network require encryption.
