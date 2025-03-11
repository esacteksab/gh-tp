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

.PHONY: clean
clean:
ifneq (,$(wildcard ./*plan.md))
	rm *plan.md
endif

ifneq (,$(wildcard ./tpplan.md))
	rm tpplan.md
endif

ifneq (,$(wildcard ./*plan.out))
	rm *plan.out
endif

ifneq (,$(wildcard ./tpplan.out))
	rm tpplan.out
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

ifneq (,$(wildcard ./dist))
	rm -rf dist

endif

ifneq (,$(wildcard ./coverage))
	rm -f coverage.*

endif


.PHONY: build
build:

	goreleaser build --clean --single-target --snapshot
	cp dist/gh-tp_linux_amd64_v1/gh-tp .

	gh ext remove tp

	gh ext install .

	-gh tp --version


.PHONY: format
format:
	gofumpt -l -w .

.PHONY: tidy
tidy:
	go mod tidy
