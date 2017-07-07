#!/bin/bash

# Copyright (c) 2016-2017 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

MODULES=$(go list ./...)
for M in $MODULES; do
    go test -covermode=count $M
done
