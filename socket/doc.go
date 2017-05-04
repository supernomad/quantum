// Copyright (c) 2016-2017 Christian Saide <Supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

/*
Package socket contains the structs and logic to create, configure, and maintain multi-queue sockets. Each socket type is represented by a struct adhering to the included socket interface, which describes a generic multi-queue socket.

Currently supported sockets:
	- UDP socket
	- DTLS socket
*/
package socket
