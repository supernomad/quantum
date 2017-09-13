########################
 The ``quantum`` manual
########################

Welcome to the ``quantum`` manual. ``quantum`` is a lightweight and WAN oriented software defined network (SDN), that is designed with security and scalability at its heart. This manual contains a full reference to ``quantum``, and indepth descriptions of its inner workings.

What is ``quantum``?
====================

The theory, design, and implementation behind ``quantum``, is all based on the premise that there is currently no other WAN oriented software defined network on the market. ``quantum`` strives to solve the same problems as technologies like `Weave <https://www.weave.works/oss/net/>`_ and `Flanneld <https://coreos.com/flannel/docs/latest/flannel-config.html>`_, but specifically focus on infrastructure that spans multiple geographic locations and span many providers from cloud based to colocations and private datacenters.

Features
--------

The high level features of ``quantum`` are simple:
  * Thrives in a high latency and distributed infrastructure.
  * Gracefully handles and manages network partitions.
  * Transparently provides powerful network level plugins to augment traffic in real time.
  * Provides a single seamless fully authenticated and secured mesh network.
  * Exposes powerful network metrics for all connected servers, down to which network queue transmitted/received a packet.

Delving Deeper
==============

.. list-table::
   :widths: auto
   :header-rows: 1
   :align: center

   * - Getting Started
     - Administration
     - Development
     - Reference
   * - :doc:`Introduction <introduction>`
     - :doc:`Configuration Options <configuration>`
     - `Godoc <https://godoc.org/github.com/supernomad/quantum/>`_
     - :doc:`Use Cases <use-cases>`
   * - :doc:`Install <install>`
     - :doc:`Guarantees/SLAs <guarantees-sla>`
     - :doc:`Requirements <dev-reqs>`
     - :doc:`Get Help <help>`
   * - :doc:`Quick Start <quick-start>`
     - :doc:`Distaster Recovery <dr>`
     - :doc:`Contributing <contribute>`
     - :doc:`Glossary <glossary>`
   * - :doc:`FAQ <faq>`
     - :doc:`Security <security>`
     - :doc:`Extending Quantum <extending>`
     -
   * -
     - :doc:`Plugins <plugins>`
     -
     -

Special Thanks
==============

I would like to reach out to a few specific OSS projects that made ``quantum`` possible.

First and foremost I would like to give a special thanks to the different software defined networks that exist, without which I would have never been inspired to make this project. So hats off to all of you.

I would also like to thank all of the authors of the following OSS libraries which ``quantum`` relies on to function, and wouldn't exist without:
  * `OpenSSL <https://www.openssl.org/>`_
  * `Etcd Client <https://github.com/coreos/etcd>`_
  * `Semver <https://github.com/coreos/go-semver/semver>`_
  * `Snappy <https://github.com/golang/snappy>`_
  * `Codec <https://github.com/ugorji/go/codec>`_
  * `Netlink <https://github.com/vishvananda/netlink>`_
  * `Netns <https://github.com/vishvananda/netns>`_
  * `Curve25519 <https://golang.org/x/crypto/curve25519>`_
  * `Pbkdf2 <https://golang.org/x/crypto/pbkdf2>`_
  * `Context <https://golang.org/x/net/context>`_
  * `Yaml <https://gopkg.in/yaml.v2>`_

.. toctree::
   :hidden:
   :maxdepth: 2

   introduction.rst
   install.rst
   quick-start.rst
   configuration.rst
   security.rst
   plugins.rst
   guarantees-sla.rst
   dr.rst
   use-cases.rst
   dev-reqs.rst
   extending.rst
   contribute.rst
   faq.rst
   glossary.rst
   help.rst
   license.rst
