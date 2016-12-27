#!/bin/bash

# Copyright (c) 2016 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

echo "Exporting GOMAXPROCS=1"
LASTGOMAXPROCS=$GOMAXPROCS
export GOMAXPROCS=1

pushd "$( cd $(dirname $0) ; pwd -P )/.." 2>&1 > /dev/null

echo "Getting deps:"
go get -u golang.org/x/tools/cmd/cover
go get -u github.com/mattn/goveralls
go get -u github.com/golang/lint/golint
go get -u github.com/GeertJohan/fgt
echo "DONE"

echo "Running go install:"
fgt go install github.com/Supernomad/quantum && echo "PASS"

echo "Running go fmt:"
fgt go fmt ./... && echo "PASS"

echo "Running go vet:"
fgt go vet ./... && echo "PASS"

echo "Running go lint:"
fgt golint ./... && echo "PASS"

echo "Running go test:"
go test -bench . -benchmem ./...

if [[ $1 == "coverage" ]]; then
	echo "Running go cover:"
	bin/coverage $2
fi

rm quantum.pid

echo "Reseting GOMAXPROCS to $LASTGOMAXPROCS"
export GOMAXPROCS=$LASTGOMAXPROCS

popd 2>&1 > /dev/null
