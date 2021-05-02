## simple makefile to log workflow
.PHONY: all test clean build install protobuf

SHELL := /bin/bash

GOFLAGS ?= $(GOFLAGS:)

export EPS_TEST = yes

EPS_TEST_SETTINGS ?= "$(shell pwd)/settings/test"

all: dep install

build:
	@go build $(GOFLAGS) ./...

dep:
	@go get ./...

install:
	@go install $(GOFLAGS) ./...

test: dep
	EPS_SETTINGS=$(EPS_TEST_SETTINGS) go test $(testargs) `go list ./...`

test-races: dep
	EPS_SETTINGS=$(EPS_TEST_SETTINGS) go test -race $(testargs) `go list ./...`

bench: dep
	EPS_SETTINGS=$(EPS_TEST_SETTINGS) go test -run=NONE -bench=. $(GOFLAGS) `go list ./... | grep -v api/`

clean:
	@go clean $(GOFLAGS) -i ./...

copyright:
	python .scripts/make_copyright_headers.py

protobuf:
	protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    protobuf/eps.proto

certs:
	(cd settings/dev/certs; ../../../.scripts/make_certs.sh)
	(cd settings/test/certs; ../../../.scripts/make_certs.sh)
