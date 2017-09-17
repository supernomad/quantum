##############
 Introduction
##############

``quantum`` is a software defined network that is written in a combination of C and Golang. It is designed with scalability and security at its core, leveraging the latest techonologies to enable seamless global scale networking.

What is ``quantum``?
====================

The theory, design, and implementation behind ``quantum``, is all based on the premise that there is currently no other WAN oriented software defined network on the market. ``quantum`` strives to solve the same problems as technologies like `Weave <https://www.weave.works/oss/net/>`_ and `Flanneld <https://coreos.com/flannel/docs/latest/flannel-config.html>`_, but specifically focus on infrastructure that spans multiple geographic locations and span many providers from cloud based to colocations and private datacenters.

Why use ``quantum``?
====================

There are a myriad of :doc:`use cases <use-cases>`, that ``quantum`` aims to solve. The primary use case that it is designed for is to operate in nonhomogeneous infratstructure that spans many geographic regions and many providers.

Features
========

The high level features of ``quantum`` are simple:
  * Thrives in a high latency and distributed infrastructure.
  * Gracefully handles and manages network partitions.
  * Transparently provides powerful network level plugins to augment traffic in real time.
  * Provides a single seamless fully authenticated and secured mesh network.
  * Exposes powerful network metrics for all connected servers, down to which network queue transmitted/received a packet.
