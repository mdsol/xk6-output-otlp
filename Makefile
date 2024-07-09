MAKEFLAGS += --silent
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)

clean:
	echo Remove old binary if it exists...
	rm -f ./bin/*

build:
	echo Build...
	xk6 build --output ./bin/k6 --with github.com/mdsol/xk6-output-otlp=.

pack:
	echo Packing...
	cd ./bin/; \
	tar -czvf k6-$(GOARCH)-$(GOOS).tar.gz ./k6; \
	rm ./k6

format:
	echo Format...
	go fmt ./...

prepare:
	echo Installing XK6...
	go install go.k6.io/xk6/cmd/xk6@latest

lint:
	echo Linter...
	golangci-lint run

.PHONY: prepare build clean format lint pack
