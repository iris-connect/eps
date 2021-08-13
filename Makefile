## simple makefile to log workflow
.PHONY: all test clean build install protobuf examples

SHELL := /bin/bash
VERSION ?= development

GOFLAGS ?= $(GOFLAGS:) -ldflags "-X 'github.com/iris-connect/eps.Version=$(VERSION)'"

export EPS_TEST = yes

EPS_TEST_SETTINGS ?= "$(shell pwd)/settings/test"

all: dep install

build:
	go build $(GOFLAGS) ./...

dep:
	@go get ./...

install: build
	go install $(GOFLAGS) ./...

test:
	EPS_SETTINGS=$(EPS_TEST_SETTINGS) go test $(testargs) `go list ./...`

test-races:
	EPS_SETTINGS=$(EPS_TEST_SETTINGS) go test -race $(testargs) `go list ./...`

bench:
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
	rm -rf settings/dev/certs/*
	rm -rf settings/test/certs/*
	(cd settings/dev/certs; ../../../.scripts/make_certs.sh)
	(cd settings/test/certs; ../../../.scripts/make_certs.sh)

sd-setup:
	.scripts/sd_setup.sh settings/dev/directory

sd-test-setup:
	.scripts/sd_setup.sh settings/test/directory

examples:
	@go build $(GOFLAGS) -tags examples ./...
	@go install $(GOFLAGS) -tags examples ./...