# quantum
[![Build Status](https://travis-ci.org/Supernomad/quantum.svg?branch=develop)](https://travis-ci.org/Supernomad/quantum) [![Coverage Status](https://coveralls.io/repos/github/Supernomad/quantum/badge.svg?branch=develop)](https://coveralls.io/github/Supernomad/quantum?branch=develop) [![Go Report Card](https://goreportcard.com/badge/github.com/Supernomad/quantum)](https://goreportcard.com/report/github.com/Supernomad/quantum) [![GoDoc](https://godoc.org/github.com/Supernomad/quantum?status.png)](https://godoc.org/github.com/Supernomad/quantum)

`quantum` is an sdn written in go with global networking, security, and auto-configuration at its heart. It leverages distributed datastores and state of the art encryption to offer fully secured end to end global networking over a single cohesive network.

## Description
`quantum` functions entirely peer to peer, and utilizes a combination of a [TUN interface](https://www.kernel.org/doc/Documentation/networking/tuntap.txt), a [UDP socket](http://www.cs.dartmouth.edu/~campbell/cs60/socketprogramming.html), [ECDHE](https://en.wikipedia.org/wiki/Elliptic_curve_Diffie%E2%80%93Hellman) with [Curve25519](https://en.wikipedia.org/wiki/Curve25519), and [AES-256-GCM](http://crypto.stackexchange.com/questions/17999/aes256-gcm-can-someone-explain-how-to-use-it-securely-ruby) to translate private bound network traffic into encrypted and authenticated public bound traffic.

#### The theory of operation
The `quantum` process starts up a multi-queue TUN interface and a series of UDP sockets. During this process a connection to one of the supported datastore backends is made, which initiates a full sync of the nodes participating in the network. When a client/server application generates a network packet bound for one of the ip addresses that is in the `quantum` network, the TUN interface will capture that packet and pass it off to the `quantum` process. Once `quantum` recieves a packet off the TUN interface it will determine what the mapping is for that packet's destination, and encrypt the data, then send it via the UDP socket to the recipeint server. The recipeint server will then determine the mapping of the sending server, and unencrypt the data, then send it via the TUN interface to the client/server application that is waiting for said data.

The backend operations are handled in background, using a series of watch handlers, refresh handlers and sync handlers. This allows for full DHCP auto determination of private ip addressing, and for it to persist through restarts.

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
