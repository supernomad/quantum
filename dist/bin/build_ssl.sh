#!/bin/bash

cd vendor/openssl

./config
make -j ${1}
