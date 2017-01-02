// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

/*
Package datastore contains the structs and logic to handle accessing and managing a backend datastore. This datastore will be responsible for managing distribution of all network mappings, and global configuration throughout the quantum network.

This package is designed to be extended and is based off of a simple interface defined in, 'device/device.go'. With this simple interface it is feasible to implement a myriad of datastore backends, from other key/value stores like https://www.consul.io to sql databases like https://www.postgresql.org/.

The basic design of the datastore as it stands is an in memory cache in the form of a golang map object, that is synchronized with the datastore in background go routines. This allows for very fast operation of the worker routines that depend on the data while still allowing for consistent results.

The data structure itself is as follows:
	Key: Private ip of the node
	Value: json serialized mapping object

	Etcd Example:
	quantum/nodes/10.99.0.1
	{
	  "machineID": "b8fc945e893cfd55dc6170b6a4f6471d5790fa279e020410f435759ba9e3f0c5",
	  "privateIP": "10.99.0.1",
	  "publicKey": "EZOUpfx4N0LvU8A9\/b5seoUSm7+sOvWr8uE7zRATijU=",
	  "ipv4": "172.18.0.2",
	  "ipv6": "fd00:dead:beef::2",
	  "port": 1099
	}
*/
package datastore
