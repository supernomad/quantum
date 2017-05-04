// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

/*
Package device contains the structs and logic to create, configure, and maintain virtual multi-queue network devices. Each network device type is represented by a struct adhering to the included Device interface, which describes a generic multi-queue virtual network device.

Currently supported devices:
	- TUN device
*/
package device
