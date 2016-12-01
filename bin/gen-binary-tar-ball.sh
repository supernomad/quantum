#!/bin/bash

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
tar cvzf quantum_${version}_linux_amd64.tar.gz quantum

echo "Cleaning up..."
rm quantum
mv quantum_${version}_linux_amd64.tar.gz $old_working_dir/

echo "Generation complete..."
echo "Archive: $old_working_dir/quantum_${version}_linux_amd64.tar.gz"
popd
