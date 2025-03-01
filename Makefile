MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
.DELETE_ON_ERROR:
.SUFFIXES:


.PHONY: build
build:

ifneq (,$(wildcard ./plan.md))
	rm plan.md
endif

ifneq (,$(wildcard .plan.out))
	rm plan.out
endif

	scripts/build-dev.sh

	gh tp
