###########
 Operation
###########

Running ``quantum`` in a production environment is very straight forward but does require some initial planning to ensure proper security and smooth operation.

Initial Setup
=============

When first setting up ``quantum`` in your environment a few key things should be kept in mind:

  * The `primary subnet <configuration.html#primary-subnet>`_ that ``quantum`` will use for private communication which defaults to ``10.99.0.0/16``.
  * The `static ip reservation <configuration.html#reserved-static-ip-subnet>`_ which defaults to ``10.99.0.0/23``.
  * The `floating ip reservation <configuration.html#reserved-floating-ip-subnet>`_ which defaults to ``10.99.2.0/23``.
  * The :doc:`networking backend <network-backends>` to utilize which defaults to ``udp`` and insecure.
  * The :doc:`plugins <plugins>` to utilize which defaults to none.
  * The `listen port <configuration.html#listen-port>`_ to utilize for transmitting and receiving ``quantum`` network traffic which defaults to ``1099``.

Public Addressing
=================

``quantum`` by default will try to determine the local servers ip addressing and if found prefer ip v6 over ip v4. This bahvior can be modified at run time by setting the following parameters:

  * The `public ip v4 address <configuration.html#public-ipv4>`_.
  * The `public ip v6 address <configuration.html#public-ipv6>`_.
  * Utilizing the `disable ip v4 flag <configuration.html#disable-public-ipv4>`_ or the `disable ip v6 flag <configuration.html#disable-public-ipv6>`_.

  NOTE: You must have one available public ip address, however this address *can* be a NAT ip as its not bound but broadcast to the rest of the cluster for communication purposes.

This public addressing will be used in combination with `the listen port <configuration.html#listen-port>`_, as the endpoint other servers will send ``quantum`` network traffic. This combination must be unique for all servers that operate in the same cluster.

Because ``quantum`` operates in a mesh fashion and there is no middle man, firewall's should set to allow sending and receiving traffic from the entire cluster of ``quantum`` enabled servers.

Rolling Restart
===============

``quantum`` is designed with the ability to upgrade itself and update configuration on the fly, by utilizing a rolling restart function. This is done with best effort to not drop any packets during the process, but on high traffic networks the possibility of the underlying kernel buffers filling is present when the restart is in progress.

To initiate a rolling restart, all that is needed is to send a ``SIGHUP`` signal to the primary process ID for ``quantum``. This process ID can always be found at the configured `pid file <configuration.html#pid-file-path>`_ which defaults to ``/var/run/quantum.pid``. Here is an example of restarting a local ``quantum`` instance:

.. code-block:: shell

    user@host1$ kill -SIGHUP $(cat /var/run/quantum.pid)

