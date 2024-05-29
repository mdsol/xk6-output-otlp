MAKEFLAGS += --silent
GOLANGCI_CONFIG ?= .golangci.yml

clean:
	rm -f ./bin/k6	

build:
	go install go.k6.io/xk6/cmd/xk6@latest
	xk6 build --output ./bin/k6 --with github.com/mdsol/xk6-output-otlp=.

format:
	go fmt ./...

.PHONY: build clean format
