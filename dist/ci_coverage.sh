#!/bin/bash

# Copyright (c) 2016-2017 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

MODE="mode: count"

echo $MODE > full-coverage.out

MODULES=$(go list ./...)
for M in $MODULES; do
    rm -f coverage.out
    go test -covermode=count -coverprofile=coverage.out $M
    [[ -f coverage.out ]] && cat coverage.out | grep -v "$MODE" >> full-coverage.out
done

gocov convert full-coverage.out | gocov-xml > coverage.xml

rm -f coverage.out
rm -f full-coverage.out
