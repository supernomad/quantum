# Copyright (c) 2016-2017 Christian Saide <supernomad>
# Licensed under the MPL-2.0, for details see https://github.com/supernomad/quantum/blob/master/LICENSE

.PHONY: all setup_ci setup_dev gen_certs build_docker ci_deps build_deps vendor_deps lib_deps compile install lint check clean release

CI=
ifdef CI
COVERAGE_ARGS=--ci
else
COVERAGE_ARGS=--sudo
endif

VERSION = $(shell go run dist/bin/get_version.go)

all: lib_deps compile

setup_ci: ci_deps build_deps vendor_deps gen_certs

setup_dev: build_deps vendor_deps gen_certs build_docker

gen_certs:
	@echo "Generating etcd certificates..."
	@dist/bin/generate-tls-test-certs.sh

build_docker:
	@echo "Building test docker container..."
	@docker-compose build

ci_deps:
	@echo "Running go get to install ci specific build dependencies..."
	@go get -u github.com/tebeka/go2xunit
	@go get -u github.com/ryancox/gobench2plot
	@go get -u github.com/axw/gocov/...
	@go get -u github.com/AlekSi/gocov-xml

build_deps:
	@echo "Running go get to install build dependencies..."
	@pip install sphinx
	@pip install sphinx_rtd_theme
	@go get -u golang.org/x/tools/cmd/cover
	@go get -u github.com/golang/lint/golint
	@go get -u github.com/client9/misspell/cmd/misspell
	@go get -u github.com/GeertJohan/fgt

vendor_deps:
	@echo "Building vendored deps..."
	@dist/bin/build_ssl.sh $(shell cat /proc/cpuinfo | grep processor | wc -l)

lib_deps:
	@echo "Running go get to install library dependencies..."
	@go get -t -v $(go list ./... | grep -v /dist/)

compile:
	@echo "Compiling quantum..."
	@go build github.com/supernomad/quantum

install:
	@echo "Installing quantum..."
	@mv quantum /usr/sbin

html:
	@go run dist/bin/config_to_json.go docs/_static/configuration.json
	@cd docs/ && $(MAKE) html

lint:
	@echo "Running linters..."
	@fgt go fmt './...'
	@fgt go vet './...'
	@fgt golint './...'
	@find . -type f -not -path "*/ssl/**/*" -and -not -path "*/vendor/**/*" -and -not -path "*/.git/*" -and -not -path "*/_build/**" | xargs fgt misspell

check: lint
	@echo "Running tests..."
	@dist/bin/coverage.sh --bench $(COVERAGE_ARGS)
	@rm -f quantum.pid

clean:
	@echo "Cleaning up..."
	@rm -rf build_output/
	@rm -f quantum
	@rm -f *_linux_amd64.tar.gz
	@rm -f *_linux_amd64.zip
	@rm -f quantum.pid
	@cd docs/ && $(MAKE) clean

release:
	@echo "Generating release tar ball, zip file, and tag..."
	@git checkout master
	@go build github.com/supernomad/quantum
	@tar czf quantum_$(VERSION)_linux_amd64.tar.gz quantum LICENSE
	@zip quantum_$(VERSION)_linux_amd64.zip quantum LICENSE
	@rm -f quantum
	@git tag -s $(VERSION) -m "quantum v$(VERSION)"
	@git push origin $(VERSION)
