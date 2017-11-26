#!/bin/bash

# Copyright (c) 2016-2017 Christian Saide <supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

mkdir -p /dev/net
mknod /dev/net/tun c 10 200
chmod 0666 /dev/net/tun

if [[ $1 == "masquerade" ]]; then
	iptables -t nat -A POSTROUTING -s 10.99.0.0/16 -o eth0 -j MASQUERADE
	shift
fi

/bin/quantum $@
