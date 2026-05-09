.PHONY: build install test lint release

build:
	go build -o bin/querywise .

install:
	go install ./...

test:
	go test ./...

lint:
	golangci-lint run

release:
	goreleaser release --snapshot --clean
