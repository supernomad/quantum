# quantum
[![Build Status](https://travis-ci.org/Supernomad/quantum.svg?branch=develop)](https://travis-ci.org/Supernomad/quantum) [![Coverage Status](https://coveralls.io/repos/github/Supernomad/quantum/badge.svg?branch=develop)](https://coveralls.io/github/Supernomad/quantum?branch=develop) [![Go Report Card](https://goreportcard.com/badge/github.com/Supernomad/quantum)](https://goreportcard.com/report/github.com/Supernomad/quantum) [![GoDoc](https://godoc.org/github.com/Supernomad/quantum?status.png)](https://godoc.org/github.com/Supernomad/quantum)

`quantum` is an sdn written in go with global networking, security, and auto-configuration at its heart. It leverages distributed datastores and state of the art encryption to offer fully secured end to end global networking over a single cohesive network.

## Description
`quantum` is designed solve the problem of having a secure private network span multiple cloud providers as well as private colocations and datacenters, that is simple to manage and configure. `quantum` allows for you to setup a fully private global network subnet, and it would run entirely peer to peer and be end to end encrypted, utilizing [ECDHE](https://en.wikipedia.org/wiki/Elliptic_curve_Diffie%E2%80%93Hellman) with [Curve25519](https://en.wikipedia.org/wiki/Curve25519), and [AES-256-GCM](http://crypto.stackexchange.com/questions/17999/aes256-gcm-can-someone-explain-how-to-use-it-securely-ruby) to ensure the network traffic is truely fully confidential and authenticated.

#### The theory of operation
The `quantum` revolves aroung three things starting up a multi-queue TUN interface, starting up a series of UDP sockets, and initializing the backend datastore. The multi-queue TUN interface is created and configured on the fly, and is torn down on application shutdown. The UDP sockets are also generated on the fly and torn down on application shutdown, with there being the a 1 to 1 mapping between TUN queues to UDP sockets. The backend datastore is managed via a series of watches and syncs that ensure that the local cache is always as up to date as possible with the master records stored in the backend.

When a client/server application generates a network packet bound for one of the private ip addresses in the `quantum` network, the TUN interface will capture that packet and pass it off to the `quantum` process. Once `quantum` recieves a packet off the TUN interface it will determine the destination server, encrypt the data with the unique key between the sender and the recipient, then send it via the UDP socket to the recipeint server. The recipeint server will then determine the mapping of the sending server, and unencrypt the data with the unique key between the sender and the recipient, then send it via the TUN interface to the client/server application that is waiting for the packet.

New nodes are added to the datastore on startup, which triggers the current members of the network, to resync with the backend and grab the new nodes configuraiton. This also works in the reverse, so as servers are removed from the network, the rest of the network will be alerted of the change and do a resync with the backend.

#### Supported Datastores
- Consul
- Etcd

#### Supported OS's
- Linux

##### Soon to be supported:
- BSD
- Darwin

## Development Requirements
Currently `quantum` development is entirely in go and BASH. Most development has been done on ubuntu server 14.04, however any recent linux distribution with the following dependancies should be sufficent.

### Dependacies
- bash
- docker
- docker-compose
- go 1.6

## Basic operation
The basic operation of `quantum` is as simple as starting a consul or etcd server instance, and starting a few `quantum` instances pointing at it.

To make it easier to rapidly test `quantum` there is an included Dockerfile and docker-compose file. Which you can utilize by running:

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
# Run a build of quantum which will ensure your system is indeed up to date.
$ bin/build.sh
# Build the container to run quantum in.
$ docker-compose Build
# Start up etcd
$ docker-compose up -d etcd
# Start up quantum
$ docker-compose up -d quantum0
$ docker-compose up -d quantum1
$ docker-compose up -d quantum2
# Check on the status of the different quantum instances
$ docker-compose logs quantum0 quantum1 quantum2
```

The above shell snippet will bring up an etcd instance and three quantum containers using the three different configuration methods. All of `quantum` configuration is 1 to 1 between cli, environment variable, and configuraiton file. The precedence of the configuration is also in that order, meaning you can easily combine configuration options in all three of the methods, and override if need be.

To do basic testing and builds run:

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ bin/build.sh
```

To do bandwidth based testing the `quantum` containers all have iperf installed. To do testing yourself, you can run something like the following:
> Be sure to check private ip's are correct, if you use a different configuration than the following.

``` shell
# Assumes the quantum instances are started
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum0 iperf3 -s -f M

# In another shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum1 iperf3 -c 10.9.0.1 -P 2 -t 50

# In another shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum2 iperf3 -c 10.9.0.1 -P 2 -t 50
```

## Useful links
