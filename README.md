# quantum
[![Build Status](https://travis-ci.org/Supernomad/quantum.svg?branch=develop)](https://travis-ci.org/Supernomad/quantum) [![Coverage Status](https://coveralls.io/repos/github/Supernomad/quantum/badge.svg?branch=develop)](https://coveralls.io/github/Supernomad/quantum?branch=develop) [![Go Report Card](https://goreportcard.com/badge/github.com/Supernomad/quantum)](https://goreportcard.com/report/github.com/Supernomad/quantum) [![GoDoc](https://godoc.org/github.com/Supernomad/quantum?status.png)](https://godoc.org/github.com/Supernomad/quantum)

A lightweight, end to end encrypted, WAN oriented sdn written entirely in go.

## Description
`quantum` is a sdn written in go with global networking, security, and auto-configuration at its heart. It leverages a configurable distributed datastore backend and local caching to offer fully end to end encrypted networking over a single cohesive network.

The networking itself takes place entirely peer to peer, and each link between peers has a unique encryption key. The encryption utilizes a combination of symetric [AES-256](https://en.wikipedia.org/wiki/Advanced_Encryption_Standard) [GCM](https://en.wikipedia.org/wiki/Galois/Counter_Mode) backed by [Curve25519](https://en.wikipedia.org/wiki/Curve25519) [ECDHE](https://en.wikipedia.org/wiki/Elliptic_curve_Diffie%E2%80%93Hellman) (Elliptic Curve Diffie–Hellman) shared secret generation. This encryption scheme enables fully confidentiality, and authentication of all network traffic over the sdn. The confidentiallity comes from AES-256, while the authentication comes from the combination of using GCM with Curv25519.

The basic theory of operation is that `quantum` creates a TUN interface and a UDP socket, any network traffic on a server bound for a private ip address within the `quantum` network will be picked up by the TUN interface, this traffic is then natted to a public ip address and port mapping, encrypted, and sent over the UDP socket to the public ip address and port mapping defined for the original private ip address. The other server will recieve this traffic via the UDP socket and then unencrypt the data before sending it to the local TUN interface, which is where the kernel will then route the traffic to the correct application that is listening on the original private ip and port mapping.

## Development Requirements

## Basic operation

## Useful links
