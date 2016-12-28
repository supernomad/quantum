PUSH_COVERAGE=""

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
	rm -f quantum.pid

release:
	@echo "Generating release tar balls..."; \
	go build github.com/Supernomad/quantum; \
	tar cvzf quantum_$(VERSION)_linux_amd64.tar.gz quantum LICENSE; \
	rm -f quantum

ci: PUSH_COVERAGE="push" build_deps deps lint compile unit coverage cleanup

local: deps lint compile bench coverage cleanup
