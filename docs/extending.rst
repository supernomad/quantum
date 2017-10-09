###################
 Extending Quantum
###################

``quantum`` has been designed to be as extendable as possible, and is written primairily using flexible and simple interfaces. There are some key places to look to start:

  * The datastore module
  * The device module
  * The metric module
  * The plugin module
  * The rest module
  * The socket module

Datastore
=========

The datastore module is used by ``quantum`` to interact with the backend key/value datastore.

Device
======

The device module handles creating and managing the life cycle of network devices used by ``quantum``. Currently only a TUN device has been implemented.

Metric
======

The metric module is where ``quantum`` aggregates the various data points that it keeps track of.

Plugin
======

The plugin module implements logic to apply to each packet that ``quantum`` interacts with. Currently there are two plugins installed ``compression`` and ``encryption``.

Rest
====

The rest module exposes a simple rest api, for users to interact with. Currently only a stats endpoint has been implemented.

Socket
======

The socket implements various socket types to use for network communication to remote hosts. Currently plain ``udp`` and ``dtls`` sockets have been implemented.
