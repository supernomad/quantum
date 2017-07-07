# `quantum`
[![Build Status](https://travis-ci.org/Supernomad/quantum.svg?branch=develop)](https://travis-ci.org/Supernomad/quantum) [![Coverage Status](https://coveralls.io/repos/github/Supernomad/quantum/badge.svg?branch=develop)](https://coveralls.io/github/Supernomad/quantum?branch=develop) [![Go Report Card](https://goreportcard.com/badge/github.com/Supernomad/quantum)](https://goreportcard.com/report/github.com/Supernomad/quantum) [![GoDoc](https://godoc.org/github.com/Supernomad/quantum?status.png)](https://godoc.org/github.com/Supernomad/quantum)

`quantum` is a software defined network device written in go with global networking, security, and auto-configuration at its heart. It leverages the latest distributed data stores and state of the art encryption to offer fully secured end to end global networking over a single cohesive network. `quantum` is fully opensource, and is licensed under the MPL-2.0, for details see [the license.](https://github.com/Supernomad/quantum/blob/master/LICENSE)

> For detailed information on the operation and configuration of `quantum` take a look at the [wiki](https://github.com/Supernomad/quantum/wiki).

### Operation
`quantum` is designed to be essentially plug and play, however the default configuration will run `quantum` in an insecure mode and assumes each node running quantum is also running its own instance of the data store. In reality for `quantum` to guarantee safe operation it **must** be run with either the `encryption` plugin enabled or using the DTLS backend. As well as having TLS fully setup and configured with the datastore.

| Supported Datastores | Supported OS's | Supported Providers |
|:----------------------------:|:--------------:|:-------------------:|
|[etcd](https://github.com/coreos/etcd)  | linux | AWS, GCE, Azure |
| | | Packet, Digital Ocean, Rackspace |
| | | Private datacenters, Private co-locations, and many more |

#### Configuration
`quantum` can be configured in any combination of three ways, cli arguments, environment variables, and configuration file entries. All configuration options are optional and have sane defaults, however runnig without parameters will force quantum to run in insecure mode. All three variants can be used in conjunction to allow for overriding variables depending on environment, the hierarchy is as follows:

- Cli parameters override all other methods
- Environment variables override configuration file entries but can be overridden by cli parameters
- Configuration file entries will override defaults but can be overridden by either cli parameters or environment variables

Run `quantum -h|--help` for a current list of configuration options or see the [wiki on configuration](https://github.com/Supernomad/quantum/wiki/Configuration) for further information.

#### Security
The security that `quantum` can guarantee is based on a few pieces of configuration. Review the following sections for a high level overview of the configuration needed to make `quantum` secure, and for a detailed overview of the different options see the [wiki on security.](https://github.com/Supernomad/quantum/wiki/Security).

##### Datastore
[Etcd security configuration](https://coreos.com/etcd/docs/latest/security.html) is very well documented and implemented. It is highly recommended to fully read and understand the security setup for etcd before considering quantum for production use. The security provided by `quantum` is intrinsicly linked to the security used by etcd, as the information stored in etcd is highly sensitive and should be kept out of sight of prying eyes.

To ensure the secure operation of `quantum` the following must be true:
- Require TLS communication to etcd in all cases.
- Each server running `quantum` has its own unique TLS client certificate/key pair.
- Etcd client certificate authentication should be enabled.
- For added security each server should also have a unique username/password to access etcd.

> For a minimalistic openssl configuration that can be used to generate test certificates see the included `dist/ssl/generate-tls-test-certs.sh` bash script

##### DTLS
The `DTLS` backend network is the most secure way to use `quantum`, this backend configures and uses DTLS v1.2 based on OpenSSL v1.1.0f. While this backend is the most secure, it also requires the most configuration to properly use. Specifically a fully configured and secured CA is needed, and each server should be given its own signed client certificate/key pair that is set to use the unique public host IP as the common name for verification purposes. The other caveat of using the `DTLS` backend network, is that all servers in the `quantum` network will use `DTLS` for communication, whether or not encryption is needed.

> Again for a minimalistic openssl configuration that can be used to generate test certificates see the included `dist/ssl/generate-tls-test-certs.sh` bash script

##### Encryption Plugin
The `Encryption Plugin` allows for secure communication using randomly generated ECDH key pairs for each server using [curve25519](https://cr.yp.to/ecdh.html). While this plugin is easier to utilize than the `DTLS` backend network it is not as secure. Due to the fact that there is no authentication of the communicating peers. However the messages that are received are authenticated using GCM guaranteeing that there is no tampering with messages between servers in transit. The `Encryption Plugin` utilizes a combination of the randomly generated ECDH key pairs, a unique random salt, pbkdf2, and AES-256-GCM. Unlike the `DTLS` backend network, only servers with this plugin enabled will communicate with encryption, which allows for granular configuraion of which servers require the security provided.

### Development
Currently `quantum` development is entirely in go and utilizes a few BASH scripts to facilitate builds and setup. Development has been mostly done on ubuntu server 14.04+, however any recent linux distribution with the following dependencies should be sufficient to develop `quantum`.

#### Development Dependencies
- bash
- make
- tun kernel module must be enabled
  - please see your distributions information on how to enable it.
- docker
- docker-compose
- openssl
- go 1.8.x
- a recent c/c++ compiler

#### Getting started
To get started developing `quantum`, run the following shell commands to get your environment configured and running.

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
# Setup the dev environment
$ make setup_dev
# Run a full development build including linting and unit tests
$ make test
# Start up the docker test bench
$ docker-compose up -d
# Wait a few seconds for initialization to complete
# Check on the status of the different quantum containers
$ docker-compose logs quantum0 quantum1 quantum2
# Run ping to ensure connectivity over quantum
$ docker exec -it quantum1 ping 10.99.0.1
$ docker exec -it quantum2 ping 10.99.0.1
```
After running the above you will have a single etcd container and three quantum containers running. The three quantum containers are configured to run a quantum network `10.99.0.0/16`, with `quantum0` having a statically defined private ip `10.99.0.1` and `quantum1`/`quantum2` having DHCP defined private ip addresses.

#### Testing
To run basic unit testing and builds execute:

``` shell
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ make test
```

To run code level benchmarks execute:

``` shell
# This must be executed as root due to the need to create a tun interface see device/device_test.go for details
$ sudo -i bash -c "cd $GOPATH/src/github.com/Supernomad/quantum; PATH='$PATH' GOPATH='$GOPATH' make bench"
```

To do basic bandwidth based testing the `quantum` containers all have iperf3 installed. For example to test how much throughput `quantum0` can handle from both `quantum1`/`quantum2`:

``` shell
# Assuming the three development quantum containers are started
# Start three shells

# In first shell start iperf3 server
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum0 iperf3 -s -f M

# In second shell start iperf3 client in quantum1
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum1 iperf3 -c 10.99.0.1 -P 2 -t 50

# In third shell start iperf3 client in quantum2
$ cd $GOPATH/src/github.com/Supernomad/quantum
$ docker exec -it quantum2 iperf3 -c 10.99.0.1 -P 2 -t 50
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
- `make dev` script must be run successfully before the PR is open.
- Documentation is added for new user facing functionality.

---
Copyright (c) 2016-2017 Christian Saide <Supernomad>
