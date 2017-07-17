#!/bin/bash

# Copyright (c) 2016-2017 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

MODE="mode: count"
SUDO="false"
CI="false"
CI_OUTPUT_DIR="build_output"
BENCH_ARGS=""
CI_ARGS=""

MODULES=$(go list ./...)

function die() {
    [[ $# -gt 0 ]] && echo "$@"
    exit 1
}

function usage() {
    cat <<EOF
Usage:
$(basename $BASH_SOURCE) [flags]

This script runs coverage analysis on the various unit and benchmark tests within quantum and its modules.

Flags:
    -s|--sudo     Whether or not to run with sudo, which is needed for unpriviledged users to create and configure TUN/TAP devices. (default: ${SUDO})
    -c|--ci       Whether this run should include ci analysis of the resulting coverage output. (default: ${CI})
    -d|--ci-dir   The directory to place CI related output in. (default: '${CI_OUTPUT_DIR}/')
    -b|--bench    Whether or not to run benchmark tests during this run. (default: false)
    -x|--debug    Print debuging statements. (default: false)
    -h|--help     Print this usage information.
EOF
    die
}

function setup() {
    echo $MODE > full-coverage.out
    touch testing_output.out

    if [[ ! -d /dev/net ]] || [[ ! -c /dev/net/tun ]]; then
        mkdir -p /dev/net
        mknod /dev/net/tun c 10 200
        chmod 0666 /dev/net/tun
    fi
}

function cleanup() {
    rm -f tmp-coverage.out
    rm -f full-coverage.out
    rm -f testing_output.out
}

function test_no_sudo() {
    go test ${CI_ARGS} -covermode=count -coverprofile=tmp-coverage.out ${BENCH_ARGS} ${1} 2>&1 \
        | tee -a testing_output.out
}

function test_with_sudo() {
    sudo -i bash -c \
        "cd $GOPATH/src/github.com/Supernomad/quantum; PATH='$PATH' GOPATH='$GOPATH' \
        go test ${CI_ARGS} -covermode=count -coverprofile=tmp-coverage.out ${BENCH_ARGS} ${1} 2>&1 \
        | tee -a testing_output.out"
}

function main() {
    for module in ${MODULES}; do
        rm -f tmp-coverage.out

        if [[ $SUDO == "true" ]]; then
            test_with_sudo ${module}
        else
            test_no_sudo ${module}
        fi

        if [[ -f tmp-coverage.out ]]; then
            grep -v "$MODE" tmp-coverage.out >> full-coverage.out
        fi
    done
}

function handle_ci() {
    rm -rf ${CI_OUTPUT_DIR}
    mkdir ${CI_OUTPUT_DIR}

    gocov convert full-coverage.out | gocov-xml > ${CI_OUTPUT_DIR}/coverage.xml
    sed -i -e "s:/opt/go/src/github.com/Supernomad/quantum/::g" ${CI_OUTPUT_DIR}/coverage.xml

    cat testing_output.out | go2xunit -output ${CI_OUTPUT_DIR}/tests.xml
    cat testing_output.out | gobench2plot > ${CI_OUTPUT_DIR}/benchmarks.xml
}


while [[ $1 ]]; do
    case "$1" in
        -s|--sudo)    SUDO="true" ;;
        -c|--ci)      CI="true"; CI_ARGS="-v" ;;
        -b|--bench)   BENCH_ARGS="-bench . -benchmem" ;;
        -x|--debug)   set -x ;;
        -h|--help|*)  usage ;;
    esac
    shift
done


cleanup
setup

if [[ $BENCH == "true" ]]; then
    BENCH_ARGS="-bench . -benchmem"
fi

main

if [[ $CI == "true" ]]; then
    handle_ci
fi

cleanup
