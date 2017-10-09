#############
 Quick Start
#############

``quantum`` by default runs in an insecure mode. These defaults facilitate ease of access and should be overriden as needed in your infrastructure. For detailed information runing within a production environment and security see the docs on :doc:`operating quantum <operation>` and :doc:`on security <security>`.

To run ``quantum``, you will need a basic etcd service and two servers that can both access your etcd service. Go ahead and :doc:`install quantum <install>` on your two servers under test, and once installed ensure that the two servers can communicate with each other via udp on port ``1099``. You can then run the following on the two servers under test to verify your installation and start up your first cluster:

    NOTE: ETCD_HOSTS should be set to one or more of the IP:PORT or HOST:PORT combinations for your etcd service.

On host1:

.. code-block:: shell

    user@host1$ quantum -h
    user@host1$ quantum --datastore-endpoints "${ETCD_HOSTS}" --datastore-prefix "/testing" --private-ip "10.99.0.1"
    [INFO] [MAIN] Listening on device:  quantum0
    [INFO] [MAIN] Network space:        10.99.0.0/16
    [INFO] [MAIN] Private IP address:   10.99.0.1
    [INFO] [MAIN] Public IPv4 address:  # this servers public ip v4 address as determined by quantum if available
    [INFO] [MAIN] Public IPv6 address:  # this servers public ip v6 address as determined by quantum if available
    [INFO] [MAIN] Listening on port:    1099
    [INFO] [MAIN] Using backend:        udp
    [INFO] [MAIN] Using plugins:

On host2:

.. code-block:: shell

    user@host2$ quantum -h
    user@host2$ quantum --datastore-endpoints "${ETCD_HOSTS}" --datastore-prefix "/testing" --private-ip "10.99.0.2"
    [INFO] [MAIN] Listening on device:  quantum0
    [INFO] [MAIN] Network space:        10.99.0.0/16
    [INFO] [MAIN] Private IP address:   10.99.0.2
    [INFO] [MAIN] Public IPv4 address:  # this servers public ip v4 address as determined by quantum if available
    [INFO] [MAIN] Public IPv6 address:  # this servers public ip v6 address as determined by quantum if available
    [INFO] [MAIN] Listening on port:    1099
    [INFO] [MAIN] Using backend:        udp
    [INFO] [MAIN] Using plugins:

Now that the servers are up and running, go ahead and start communicating using the private ip addresses. Once you have transmitted some data take a look at the metrics that ``quantum`` collected:

On host1:

.. code-block:: shell

    user@host1$ ping 10.99.0.2 -c 5
    # Then check the stats locally and on the remote server.
    user@host1$ curl localhost:1099/stats?pretty
    user@host1$ curl 10.99.0.2:1099/stats?pretty
