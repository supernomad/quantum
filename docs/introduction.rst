##############
 Introduction
##############

``quantum`` is a software defined network that is written in a combination of C and Golang. It is designed with scalability and security at its core, leveraging the latest techonologies to enable seamless global scale networking.

What makes ``quantum`` different?
=================================

The theory, design, and implementation behind ``quantum``, is all based on the premise that there is currently no other WAN oriented software defined network on the market. ``quantum`` strives to solve the same problems as technologies like `Weave <https://www.weave.works/oss/net/>`_ and `Flanneld <https://coreos.com/flannel/docs/latest/flannel-config.html>`_, but specifically focus on infrastructure that spans multiple geographic locations and span many providers from cloud based to colocations and private datacenters.

There is also one huge difference between ``quantum`` and the other virtual networking systems that exist today. That difference is that it operates on the **host** level rather than the **container** level. The reasoning behind this difference is rather simple, and its because networking between services on a single host can and should be handled by the OS. Every OS has a virtual network built in and exposed on the loopback interface that has a full `class A IP v4 subnet <https://en.wikipedia.org/wiki/Classful_network>`_ at its disposal, ``127.0.0.0/8``, and has the ability add any amount of `IP v6 unique local addresses <https://en.wikipedia.org/wiki/Unique_local_address>`_, from ``fc00::/7``. It is the opinion of ``quantum`` that there is no need to add a layer of abstraction ontop of this hardned OS capability. So instead of worrying about linking containers together on the same host, ``quantum`` focuses on linking hosts together.

This has a two major benefits:

  * The overall size and complexity of the ``quantum`` network is drastically cut down.
  * There is a much higher upper limit to the number of containers that can be run in a single network.
    * Instead of having a ``/16`` for all hosts **and** containers, you can have a ``/16`` for hosts and a ``/8`` **per** host number of conatiners all with unique addressing.

The major drawback is that having a hostname per container does not work with ``quantum``, however this is unlikely to be necesary in real production deployments. Where you are likely to have hundreds if not thousands of containers running. The thought here is that its just as easy to remeber ``hostA.example.com`` has api services running on the port range ``1000:2000`` as it is to remember ``hostA.api0001.example.com``.

Why use ``quantum``?
====================

There are a myriad of :doc:`use cases <use-cases>`, that ``quantum`` aims to solve. In order to determine whether or not ``quantum`` is right for your use case please ask these key questions:

  * Is your infrastructure spread accross many different providers?
  * Is your infrastructure spread over many geographic regions?
  * Would you benefit from detailed and highly granular network level metrics?
  * Do you want to easily apply network level operations such as compression, encryption, authentication to your traffic on a per node basis?
  * Do you care more about server to server communication than inter container communication?

If the answer to any of the above questions is "Yes", then ``quantum`` may be a very good candidate for your infrastructure. However if the answer to the majority of the above questions it might be better to consider the other amazing technology like `Weave <https://www.weave.works/oss/net/>`_ and `Flanneld <https://coreos.com/flannel/docs/latest/flannel-config.html>`_.
