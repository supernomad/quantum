#!/bin/bash

if [[ -z $1 ]]; then
    echo "Usage: $0 [version]"
    exit 1
fi

old_working_dir="$(pwd)"
version=$1
pushd /tmp/

echo "Building quantum"
go build github.com/Supernomad/quantum
tar cvzf quantum_0.11.0_linux_amd64.tar.gz quantum
rm quantum

mv quantum_${version}_linux_amd64.tar.gz $old_working_dir/

popd
