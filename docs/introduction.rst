##############
 Introduction
##############

``quantum`` is a software defined network that is written in a combination of C and Golang. It is designed with scalability and security at its core, leveraging the latest techonologies to enable seamless global scale networking.

What makes ``quantum`` different?
=================================

The theory, design, and implementation behind ``quantum``, is all based on the premise that there is currently no other WAN oriented software defined network on the market. ``quantum`` strives to solve the same problems as technologies like `Weave <https://www.weave.works/oss/net/>`_ and `Flanneld <https://coreos.com/flannel/docs/latest/flannel-config.html>`_, but specifically focus on infrastructure that spans multiple geographic locations, as well as many providers from cloud based to colocations and private datacenters.

There is also one huge difference between ``quantum`` and the other virtual networking systems that exist today. That difference is that it operates on the **host** level rather than the **container** level.

This has a few major benefits:

  * The overall complexity of the ``quantum`` network is drastically reduced.
  * The scalability of the network is greatly increased, as ip's are only handed out to servers as opposed to the potentially large number of containers.
  * Can be integrated with existing local software defined networking easily.

Why use ``quantum``?
====================

There are a myriad of :doc:`use cases <use-cases>`, that ``quantum`` aims to solve. In order to determine whether or not ``quantum`` will fit your use case please ask yourself these key questions:

  * Is your infrastructure spread accros many different providers?
  * Is your infrastructure spread over many geographic regions?
  * Would you benefit from detailed and highly granular network level metrics?
  * Do you want to easily apply network level operations such as compression, encryption, and authentication to your traffic on a per peer basis?
  * Do you care more about server to server communication than inter container communication?

If the answer to any of the above questions is "Yes", then ``quantum`` may be a very good candidate for your infrastructure.
