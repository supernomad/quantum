############################
 Frequently Asked Questions
############################

Q. How do I determine the local ``quantum`` ip address is assigned to a server?
    You can always check the current assigned ip address by checking the ``QUANTUM_IP`` environment variable.

Q. I can't communicate with a specific remote node?
    You should confirm that UDP traffic is allowed between the nodes `listen-port <configuration.html#listen-ip>`_'s and public ip addressing.

Q. Why am I seeing dropped packets in the ``quantum`` statistics?
    This can happen for various reasons, most notably because there is a server attempting to communicate with a secured cluster that is not authenticated.
