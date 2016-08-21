# quantum
[![Build Status](https://travis-ci.org/Supernomad/quantum.svg?branch=develop)](https://travis-ci.org/Supernomad/quantum) [![Coverage Status](https://coveralls.io/repos/github/Supernomad/quantum/badge.svg?branch=develop)](https://coveralls.io/github/Supernomad/quantum?branch=develop) [![Go Report Card](https://goreportcard.com/badge/github.com/Supernomad/quantum)](https://goreportcard.com/report/github.com/Supernomad/quantum) [![GoDoc](https://godoc.org/github.com/Supernomad/quantum?status.png)](https://godoc.org/github.com/Supernomad/quantum)

## Description
`quantum` is an sdn written in go with global networking, security, and auto-configuration at its heart. It leverages distributed datastores and state of the art encryption to offer fully secured end to end global networking over a single cohesive network.

- Encrypted with [AES-256-GCM](http://crypto.stackexchange.com/questions/17999/aes256-gcm-can-someone-explain-how-to-use-it-securely-ruby).
  - Ensuring both confidentiality of all network traffic but also authentication of both the recipient and sender.
- Secret generation using [ECDHE](https://en.wikipedia.org/wiki/Elliptic_curve_Diffie%E2%80%93Hellman) with [Curve25519](https://en.wikipedia.org/wiki/Curve25519).
  - Ensuring secret generation/transmission is always as secure as possible.
- Fully peer to peer communication.
  - Minimizing bottlenecks and maximizing performance.
- Lightweight and designed to run on systems with limited available resources.
- Designed with global distributions in mind.
  - Ability to scale to thousands of nodes spanning any geographic region.

#### The theory of operation
The `quantum` application revolves around three things, starting up a multi-queue TUN interface, starting up a series of UDP sockets, and initializing the backend datastore. The multi-queue TUN interface is created and configured on the fly, and is torn down on application shutdown. There is one UDP sockets generated for each TUN queue generated, creating a one to one mapping between the TUN interface and the UDP sockets. The backend datastore is managed via a series of watches and syncs that ensure that the local cache is always, as up to date as possible with the master records stored in the backend.

The TUN interface interacts with the kernel reading packets that are sent to a private network that its assigned. The UDP sockets are utilized to facilitate sending the encrypted packets to the public address of the destination. So when a client/server application generates a network packet bound for one of the private ip addresses in the `quantum` network, the TUN interface will capture that packet and pass it off to the `quantum` process. Once `quantum` recieves a packet off the TUN interface it will determine the destination server, encrypt the data with the unique key between the sender and the recipient, then send it via the UDP socket to the recipeint server. The recipeint server will then determine the mapping of the sending server, and unencrypt the data with the unique key between the sender and the recipient, then send it via the TUN interface to the client/server application that is waiting for the packet.

When new nodes are added to the network, each will add themselves as a mapping to the datastore on startup, which triggers the current members of the network to resync the mapping data.

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

For more information on operating quantum in a staging/production setting please see the [wiki](https://github.com/Supernomad/quantum/wiki).

To get started developing `quantum`, run the following shell commands to get your environment up and running.

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
# Get build dependancies
$ go get -t -v ./...
$ go get golang.org/x/tools/cmd/cover
$ go get github.com/mattn/goveralls
$ go get github.com/golang/lint/golint
$ go get github.com/GeertJohan/fgt
# Run a build of quantum which will ensure your system is indeed up to date.
$ bin/build.sh
# Build the container to run quantum in.
$ docker-compose build
# Start up etcd
$ docker-compose up -d etcd
# Start up quantum
$ docker-compose up -d quantum0
$ docker-compose up -d quantum1
$ docker-compose up -d quantum2
# Check on the status of the different quantum instances
$ docker-compose logs quantum0 quantum1 quantum2
```
After running the above you will have a single etcd instance and three quantum instances running. The three quantum instances are configured to run a quantum network `10.9.0.0/16`, with `quantum0` having a statically defined private ip `10.9.0.1` and `quantum1`/`quantum2` having DHCP defined private ip addresses.

To run basic unit testing and builds run:

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ bin/build.sh
```

To do bandwidth based testing the `quantum` containers all have iperf3 installed. For example to test how much through put quantum0 can handle from both quantum1/quantum2

``` shell
# Assumeing the three development quantum instances are started
# Start three shells

# In first shell start iperf3 server
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum0 iperf3 -s -f M

# In second shell start iperf3 client in quantum1
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum1 iperf3 -c 10.9.0.1 -P 2 -t 50

# In third shell start iperf3 client in quantum2
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum2 iperf3 -c 10.9.0.1 -P 2 -t 50
```

## Useful links
