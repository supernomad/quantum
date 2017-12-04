// Copyright (c) 2016-2017 Christian Saide <supernomad>
// Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

package netlink

import (
	"net"
)

const (
	testV4 = net.ParseIP("8.8.8.8")
	testV6 = net.ParseIP("2001:4860:4860::8888")
)
