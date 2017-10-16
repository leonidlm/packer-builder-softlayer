SHELL := /bin/bash
VERSION ?= $(shell git describe --tags --always --dirty)
RELEASE_NAME := packer-builder-softlayer

bootstrap:
	./script/bootstrap

.PHONY: build
build:
	export RELEASE_NAME=packer-builder-softlayer; ./script/build

.PHONY: test
test:
	./script/test
