// Copyright (c) 2016-2018 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

/*
Package device contains the structs and logic to create, configure, and maintain virtual multi-queue network devices. Each network device type is represented by a struct adhering to the included Device interface, which describes a generic multi-queue virtual network device.

Currently supported devices:
	- TUN device
*/
package device
