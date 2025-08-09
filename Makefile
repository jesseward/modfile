.PHONY: all
all: setup fmt build test

.PHONY: setup
setup:
	go mod download
	go mod tidy

.PHONY: build
build: clean
	@echo "==> Building Packages <=="
	go build -v ./...
	@echo "==> Building cli <=="
	go build -o dist/cli ./cmd/impulse/...

.PHONY: test
test:
	@echo "==> running Go tests <=="
	go test -race ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	@echo "==> Cleaning dist/ <=="
	rm -fr dist/*
