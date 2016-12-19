#!/bin/bash

# Copyright (c) 2016 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

if [[ -z $1 ]]; then
    echo "Usage: $0 [version]"
    exit 1
fi

old_working_dir="$(pwd)"
version=$1
pushd /tmp/

echo "Building quantum..."
go build github.com/Supernomad/quantum

echo "Creating archive..."
cp $GOPATH/src/github.com/Supernomad/quantum/LICENSE LICENSE
tar cvzf quantum_${version}_linux_amd64.tar.gz quantum LICENSE

echo "Cleaning up..."
rm quantum LICENSE
mv quantum_${version}_linux_amd64.tar.gz $old_working_dir/

echo "Generation complete..."
echo "Archive: $old_working_dir/quantum_${version}_linux_amd64.tar.gz"
popd
