#!/bin/bash

# Copyright (c) 2016 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

cwdir=$PWD

rm -r $GOPATH/src/github.com/Supernomad/quantum/bin/certs
mkdir -p $GOPATH/src/github.com/Supernomad/quantum/bin/certs

cd $GOPATH/src/github.com/Supernomad/quantum/bin

touch certs/index.txt
echo '01' > certs/serial

export SAN="IP:127.0.0.1, IP:172.18.0.4, IP:fd00:dead:beef::4"

# Create CA cert
openssl req -config openssl.cnf -new -x509 -extensions v3_ca -keyout certs/ca.key -out certs/ca.crt -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=ca.quantum.dev"

# Create and sign etcd server certificate
openssl req -config openssl.cnf -new -nodes -keyout certs/etcd.quantum.dev.key -out certs/etcd.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=etcd.quantum.dev"
openssl ca -config openssl.cnf -extensions etcd_server -keyfile certs/ca.key -cert certs/ca.crt -out certs/etcd.quantum.dev.crt -infiles certs/etcd.quantum.dev.csr

# Create and sign etcd client certificates
openssl req -config openssl.cnf -new -nodes -keyout certs/quantum0.quantum.dev.key -out certs/quantum0.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=quantum0.quantum.dev"
openssl ca -config openssl.cnf -extensions etcd_client -keyfile certs/ca.key -cert certs/ca.crt -out certs/quantum0.quantum.dev.crt -infiles certs/quantum0.quantum.dev.csr

openssl req -config openssl.cnf -new -nodes -keyout certs/quantum1.quantum.dev.key -out certs/quantum1.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=quantum1.quantum.dev"
openssl ca -config openssl.cnf -extensions etcd_client -keyfile certs/ca.key -cert certs/ca.crt -out certs/quantum1.quantum.dev.crt -infiles certs/quantum1.quantum.dev.csr

openssl req -config openssl.cnf -new -nodes -keyout certs/quantum2.quantum.dev.key -out certs/quantum2.quantum.dev.csr -subj "/C=US/ST=New York/L=New York City/O=quantum/OU=development/CN=quantum2.quantum.dev"
openssl ca -config openssl.cnf -extensions etcd_client -keyfile certs/ca.key -cert certs/ca.crt -out certs/quantum2.quantum.dev.crt -infiles certs/quantum2.quantum.dev.csr

cd $cwdir
