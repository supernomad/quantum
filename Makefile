# Copyright (c) 2016 Christian Saide <Supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/Supernomad/quantum/blob/master/LICENSE

PUSH_COVERAGE=""

setup_dev: build_deps gen_certs gen_docker_network build_docker

gen_certs:
	@echo "Generating etcd certificates..."; \
	dist/ssl/generate-tls-test-certs.sh

gen_docker_network:
	@echo "Setting up docker networks..."; \
	docker network create --subnet=172.18.0.0/24 --gateway=172.18.0.1 quantum_dev_net_v4 && \
	docker network create --subnet='fd00:dead:beef::/64' --gateway='fd00:dead:beef::1' --ipv6 quantum_dev_net_v6

rm_docker_network:
	@echo "Removing docker networks..."; \
	docker network rm quantum_dev_net_v4 quantum_dev_net_v6 || true

build_docker:
	@echo "Building test docker container..."; \
	docker-compose build

compile:
	@echo "Compiling quantum..."; \
	go install github.com/Supernomad/quantum

build_deps:
	@echo "Running go get to install build dependancies..."; \
	go get -u golang.org/x/tools/cmd/cover && \
	go get -u github.com/mattn/goveralls && \
	go get -u github.com/golang/lint/golint && \
	go get -u github.com/GeertJohan/fgt

deps:
	@echo "Running go get to install library dependancies..."; \
	go get -t -v './...'

lint:
	@echo "Running fmt/vet/lint"; \
	fgt go fmt './...' && \
	fgt go vet './...' && \
	fgt golint './...'

bench:
	@echo "Running tests with benchmarking enabled..."; \
	go test -bench . -benchmem './...'

unit:
	@echo "Running tests with benchmarking disabled..."; \
	go test './...'

coverage:
	@echo "Running go cover..."; \
	dist/coverage.sh $(PUSH_COVERAGE)

cleanup:
	@echo "Cleaning up..."; \
	rm -f quantum.pid

release:
	@echo "Generating release tar balls..."; \
	go build github.com/Supernomad/quantum; \
	tar cvzf quantum_$(VERSION)_linux_amd64.tar.gz quantum LICENSE; \
	rm -f quantum

travis: PUSH_COVERAGE="push" build_deps deps lint compile unit coverage

dev: deps lint compile unit cleanup
