#########
 Install
#########

To start using ``quantum`` you have two different installation options. You can download a prebuilt binary, or you can build from source.

First off there are some prerequisites:

  * At least version 3.x kernel or higher.
  * The TUN module enabled and operational.

Prebuilt Binary
===============

Navigate to `quantum releases <https://github.com/supernomad/quantum/releases>`_ and pick one of the available versions and either the ``.tar.gz`` or ``.zip`` for your platform.

.. code-block:: console

   $ tar -C /usr/sbin/ -xzf quantum_${QUANTUM_VERSION}_${QUANTUM_PLATFORM}_${QUANTUM_ARCH}.tar.gz --exclude='LICENSE'
   $ chmod +x /usr/sbin/quantum
   -- OR --
   $ unzip quantum_${QUANTUM_VERSION}_${QUANTUM_PLATFORM}_${QUANTUM_ARCH}.zip -x LICENSE -d /usr/sbin/
   $ chmod +x /usr/sbin/quantum

  Note: You may need to run the above commands with 'sudo'.

To run ``quantum`` as a service please see :ref:`running_as_a_serivce`.

Build
=====

In order to build ``quantum`` from source, you will need a few dependancies.

  * git
  * go 1.9.x
  * GNUMake
  * any recent C compiler

Once the dependancies are installed run the following:

.. code-block:: console

   $ git clone https://github.com/supernomad/quantum.git
   $ cd quantum/
   $ git checkout ${QUANTUM_VERSION}
   $ git submodule update --init
   $ make build_deps vendor_deps
   $ make
   $ make install

  Note: You may need to run 'make install' with 'sudo'.

.. _running_as_a_serivce:

Running as a Service
====================

In production, ``quantum`` should be run as a service set to start when the hosts networking starts. In order to facilitate rolling restarts, ``quantum`` exposes a PID file which will always contain the PID of the running master process. This means that your init systems needs to watch this PID and not the PID of the process it originally started to properly track its state through a rolling restart. See the following examples of how to do this with various popular init systems.

Systemd
-------

To run ``quantum`` in systemd see `this example unit file <https://github.com/supernomad/quantum/blob/master/dist/systemd/quantum.service>`_.

Upstart
-------

To run ``quantum`` in upstart see `this example configuration file <https://github.com/supernomad/quantum/blob/master/dist/upstart/quantum.conf>`_.
