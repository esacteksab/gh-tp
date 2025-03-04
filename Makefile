MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: audit
audit:
	go vet ./...
	go tool -modfile=go.tool.mod staticcheck ./...
	go tool -modfile=go.tool.mod govulncheck ./...
	golangci-lint run -v
# docker run -t --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.64.5 golangci-lint run -v

.PHONY: clean
clean:
ifneq (,$(wildcard ./plan.md))
	rm plan.md
endif

ifneq (,$(wildcard ./plan.out))
	rm plan.out
endif

ifneq (,$(wildcard ./*.tf))
	rm *.tf
endif

ifneq (,$(wildcard ./*.tofu))
	rm *.tofu
endif
ifneq (,$(wildcard ./gh-tp))
	rm gh-tp
endif

	rm -rf dist
	rm -f coverage.*


.PHONY: build
build:

ifneq (,$(wildcard ./plan.md))
	rm plan.md
endif

ifneq (,$(wildcard ./plan.out))
	rm plan.out
endif

	# scripts/build-dev.sh
	goreleaser build --clean --single-target --snapshot
	cp dist/gh-tp_linux_amd64_v1/gh-tp .

	gh ext remove tp

	gh ext install .

	-gh tp



.PHONY: mod
mod:
	go mod tidy
