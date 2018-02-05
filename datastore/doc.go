// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

/*
Package datastore contains the structs and logic to handle accessing and managing a backend datastore. The datastore contains all network mappings for the nodes participating in the quantum network, as well as global configuration for the network.

The design of the datastore module is to expose a single method that represents accessing a network mapping. This is wrapped in a simple interface to allow for extending quantum to support multiple backends in the future.

The basic architecture is to have an in memory map object that is synchronized in the background. This allows the read only worker threads efficient access to the data, while still ensuring data consistency.

Currently supported datastores:
	https://github.com/coreos/etcd

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
