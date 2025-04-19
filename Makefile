MAKEFLAGS += --warn-undefined-variables
SHELL := bash
.SHELLFLAGS := -eu -o pipefail -c
.DEFAULT_GOAL := all
.DELETE_ON_ERROR:
.SUFFIXES:

.PHONY: audit
audit: tidy fmt
	go vet ./...
	go tool -modfile=go.tool.mod staticcheck ./...
	go tool -modfile=go.tool.mod govulncheck ./...
	golangci-lint run -v


.PHONY: build
build:

	goreleaser build --clean --single-target --snapshot
	cp dist/gh-tp_linux_amd64_v1/gh-tp .

	gh ext remove tp

	gh ext install .

	-gh tp --version

.PHONY: clean
clean:
ifneq (,$(wildcard ./*plan.md))
	rm *plan.md
endif

ifneq (,$(wildcard ./*plan.out))
	rm *plan.out
endif

ifneq (,$(wildcard ./*.tf))
	rm *.tf
endif

ifneq (,$(wildcard ./*.tfstate))
	rm *.tfstate
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

.PHONY: container
container: tidy
	./scripts/build-container.sh

.PHONY: fmt
fmt:
	golines --base-formatter=gofumpt -w .
	go tool -modfile=go.tool.mod gofumpt -l -w -extra .

.PHONY: lint
lint:
	golangci-lint run -v

.PHONY: moreclean
moreclean: clean

ifneq (,$(wildcard ./.tp.toml*))
	rm .tp.toml*
endif

ifneq (,$(wildcard ~/.tp.toml*))
	rm ~/.tp.toml*
endif

ifneq (,$(wildcard ~/.config/*tp.toml*))
	rm ~/.config/*tp.toml*
endif

ifneq (,$(wildcard /var/tmp/*tp.toml*))
	rm /var/tmp/*tp.toml*
endif

ifneq (,$(wildcard ~/.config/gh-tp/.tp.toml*))
	rm -rf ~/.config/gh-tp
endif

.PHONY: test
test: container
	@mkdir -p coverdata
	@docker run --rm -v $(PWD)/coverdata:/app/coverdata esacteksab/tpt:$(shell cat .current-tag)
	@if [ -f "coverdata/coverage.out" ]; then \
		go tool cover -html=coverdata/coverage.out -o coverdata/coverage.html; \
		echo "Coverage report generated: coverage.html"; \
	else \
		echo "Error: Coverage file not found"; \
		exit 1; \
	fi

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: update
update:
	go get -u ./...
	go mod tidy
