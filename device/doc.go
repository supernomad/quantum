// Copyright (c) 2016 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

/*
Package device contains the structs and logic to create, configure, and maintain virutal network devices. Each network device type is represented by a struct adhearing to the included Device interface, which describes a generic multi-queue virutal network device.

Currently supported datastores:
	- TUN device
*/
package device
