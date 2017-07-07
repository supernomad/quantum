# Copyright (c) 2016-2017 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

BENCH_MAX_PROCS=1

test: deps lint unit bench coverage cleanup

setup_dev: build_deps vendor_deps gen_certs gen_docker_network build_docker

test_ci: deps ci_unit ci_bench ci_coverage cleanup

setup_ci: ci_deps build_deps vendor_deps gen_certs

gen_certs:
	@echo "Generating etcd certificates..."
	@dist/generate-tls-test-certs.sh

gen_docker_network:
	@echo "Setting up docker networks..."
	@docker network create --subnet=172.18.0.0/24 --gateway=172.18.0.1 quantum_dev_net_v4
	@docker network create --subnet='fd00:dead:beef::/64' --gateway='fd00:dead:beef::1' --ipv6 quantum_dev_net_v6

rm_docker_network:
	@echo "Removing docker networks..."
	@docker network rm quantum_dev_net_v4 quantum_dev_net_v6 || true

build_docker:
	@echo "Building test docker container..."
	@docker-compose build

compile:
	@echo "Compiling quantum..."
	@go install github.com/Supernomad/quantum

ci_deps:
	@echo "Running go get to install ci specific build dependencies..."
	@go get -u github.com/tebeka/go2xunit
	@go get -u github.com/ryancox/gobench2plot
	@go get -u github.com/axw/gocov/...
	@go get -u github.com/AlekSi/gocov-xml

build_deps:
	@echo "Running go get to install build dependencies..."
	@go get -u golang.org/x/tools/cmd/cover
	@go get -u github.com/mattn/goveralls
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/client9/misspell/cmd/misspell
	@go get -u github.com/GeertJohan/fgt

vendor_deps:
	@echo "Building vendored deps..."
	@(cd vendor/openssl && ./config && make && cd ../../)

deps:
	@echo "Running go get to install library dependencies..."
	@go get -t -v './...'

lint:
	@echo "Running fmt/vet/lint..."
	@fgt go fmt './...'
	@fgt go vet './...'
	@fgt golint './...'
	@find . -type f -not -path "*/ssl/**/*" -and -not -path "*/vendor/**/*" | xargs fgt misspell

bench:
	@echo "Running unit tests with benchmarking enabled..."
	@GOMAXPROCS=$(BENCH_MAX_PROCS) go test -bench=Bench* -benchmem './...'

ci_bench:
	@echo "Running unit tests with benchmarking enabled..."
	@GOMAXPROCS=$(BENCH_MAX_PROCS) go test -bench=Bench* -benchmem './...' | gobench2plot > benchmarks.xml

unit:
	@echo "Running unit tests with benchmarking disabled..."
	@go test './...'

ci_unit:
	@echo "Running ci unit tests with benchmarking disabled..."
	@go test './...' -v | go2xunit -output tests.xml

coverage:
	@echo "Running go cover..."
	@dist/coverage.sh

ci_coverage:
	@echo "Running ci coverage..."
	@dist/ci_coverage.sh

cleanup:
	@echo "Cleaning up..."
	@rm -f quantum.pid

release:
	@echo "Generating release tar ball, zip file, and tag..."
	@git checkout master
	@go build github.com/Supernomad/quantum
	@tar czf quantum_$(VERSION)_linux_amd64.tar.gz quantum LICENSE
	@zip quantum_$(VERSION)_linux_amd64.zip quantum LICENSE
	@rm -f quantum
	@git tag -s $(VERSION) -m "quantum v$(VERSION)"
	@git push origin $(VERSION)
