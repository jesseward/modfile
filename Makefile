.PHONY: all
all: setup fmt build test

.PHONY: setup
setup:
	go mod download
	go mod tidy

.PHONY: build
build: clean
	go build -v -o dist/cli cmd/pt/main.go

.PHONY: test
test:
	@echo "==> running Go tests <=="
	go test -race ./...


.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: clean
clean:
	rm -fr dist/*