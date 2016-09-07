## quantum
[![Build Status](https://travis-ci.org/Supernomad/quantum.svg?branch=develop)](https://travis-ci.org/Supernomad/quantum) [![Coverage Status](https://coveralls.io/repos/github/Supernomad/quantum/badge.svg?branch=develop)](https://coveralls.io/github/Supernomad/quantum?branch=develop) [![Go Report Card](https://goreportcard.com/badge/github.com/Supernomad/quantum)](https://goreportcard.com/report/github.com/Supernomad/quantum) [![GoDoc](https://godoc.org/github.com/Supernomad/quantum?status.png)](https://godoc.org/github.com/Supernomad/quantum)

`quantum` is a software defined network device written in go with global networking, security, and auto-configuration at its heart. It leverages the latest distributed data stores and state of the art encryption to offer fully secured end to end global networking over a single cohesive network.

> For more detailed information on the operation and configuration of `quantum` take a look at the [wiki](https://github.com/Supernomad/quantum/wiki).

### Running
`quantum` is designed to be plug and play, however the default configuration will run `quantum` in insecure mode and assumes each node has its own data store. In reality `quantum` _needs_ to be run with TLS fully setup and configured, in order to guarantee safe operation. In general it is best practice to obtain client and server certificates signed by a legitmate authority. However this does not preclude you from running with self signed certificates, as long as those certificates are kept safe and handled correctly.

| Supported Backend Datastores | Supported OS's | Supported Providers |
|:------------------:|:----:|:---------:|
|[consul](https://consul.io)  | linux | AWS, GCE, Azure |
|[etcd](https://github.com/coreos/etcd)  | | Packet, Digital Ocean, Rackspace |
| | | Private datacenters, Private co-locations, and many more |

#### Configuration
`quantum` can be configured in any combination of four ways, cli arguments, environment variables, configuration file entries, and finally falling back to defaults. Regardless of which way `quantum` is configured all of the configuration options are available.

Run `quantum -h|--help` for a current list of configuration options or see the [wiki on configuration](https://github.com/Supernomad/quantum/wiki/Configuration) for further information.

### Development
Currently `quantum` development is entirely in go and utilizes a few BASH scripts to facilitate builds and setup. Development has been mostly done on ubuntu server 14.04, however any recent linux distribution with the following dependencies should be sufficient to develop `quantum`.

#### Development Dependencies
- bash
- tun kernel module must be enabled
  - please see your distributions information on how to enable it.
- docker
- docker-compose
- go 1.6

#### Getting started
To get started developing `quantum`, run the following shell commands to get your environment configured and running.

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
# Get build dependencies
$ go get -t -v ./...
$ go get golang.org/x/tools/cmd/cover
$ go get github.com/mattn/goveralls
$ go get github.com/golang/lint/golint
$ go get github.com/GeertJohan/fgt
# Run a build of quantum which will ensure your system is indeed up to date.
$ bin/build.sh
# Build the container to run quantum in.
$ docker-compose build
# Start up docker test bench
$ docker-compose up -d
# Wait a few seconds for initialization to complete
# Check on the status of the different quantum instances
$ docker-compose logs quantum0 quantum1 quantum2
```
After running the above you will have a single etcd instance and three quantum instances running. The three quantum instances are configured to run a quantum network `10.9.0.0/16`, with `quantum0` having a statically defined private ip `10.9.0.1` and `quantum1`/`quantum2` having DHCP defined private ip addresses.

> A note about TLS, there are included TLS certificates in this repo that should only used for testing quantum in a lab setting. These certificates should not be in anyway considered secure. There is an included script `generate-etcd-certs` that was used to create the certificates, if you wish to replicate them.

#### Testing
To run basic unit testing and builds run:

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ bin/build.sh
```

To do basic bandwidth based testing the `quantum` containers all have iperf3 installed. For example to test how much through put `quantum0` can handle from both `quantum1`/`quantum2`:

``` shell
# Assuming the three development quantum instances are started
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

### Contributing
Contributions are definitely welcome, if you are looking for something to contribute check out the current [road map](https://github.com/Supernomad/quantum/milestones) and grab an open issue in the next release.

Work flow:

- Fork `quantum` from develop
- Make your changes
- Open a PR against `quantum` on to develop
- Get your PR merged
- Rinse and Repeat

There are a few rules:

- All travis builds must successfully complete before a PR will be considered.
  - Changes to travis to get builds working are ok, if they are within reason.
- The `bin/build.sh` script must be run before the PR is open.
- Documentation is added for new user facing functionality

> An aside any PR can be closed with or without explanation or justification.
