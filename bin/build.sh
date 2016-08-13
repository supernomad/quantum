#!/bin/bash
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
