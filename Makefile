SHELL := /bin/bash

INPUT_PATH ?= ./testdata/sample_chat.txt
OUTPUT_DIR ?= ./docs/data/v1
PAGES_DIR ?= ./docs
WEB_DIR ?= ./web
VERSION ?= v0.2.0
GO_PACKAGES := ./cmd/... ./internal/...

.PHONY: help install-hooks dev build data test test-integration smoke lint fmt pages-preview release clean hooks-pre-commit hooks-commit-msg hooks-pre-push hooks-post-merge hooks-post-checkout

help:
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z0-9_-]+:.*##/ {printf "%-22s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

install-hooks: ## wire local git hooks
	git config core.hooksPath .githooks
	chmod +x .githooks/*

dev: ## run frontend dev server
	npm --prefix $(WEB_DIR) run dev

data: ## regenerate static data artifacts
	go run ./cmd/build-index --input_path $(INPUT_PATH) --output_dir $(OUTPUT_DIR)

build: ## build Pages-ready frontend into docs/
	rm -rf $(PAGES_DIR)/assets
	npm --prefix $(WEB_DIR) run build
	test -f $(PAGES_DIR)/index.html
	test -f $(PAGES_DIR)/404.html
	test -s $(PAGES_DIR)/index.html

test: ## run unit tests
	go test $(GO_PACKAGES)
	npm --prefix $(WEB_DIR) run test

test-integration: ## run integration tests
	go test -tags=integration $(GO_PACKAGES)

smoke: ## build, serve docs/, and run Playwright smoke
	bash ./scripts/smoke.sh

lint: ## run linters and type checks
	gofmt -w cmd internal
	go vet $(GO_PACKAGES)
	if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run $(GO_PACKAGES); else echo "golangci-lint not found; skipping"; fi
	npm --prefix $(WEB_DIR) run lint
	npm --prefix $(WEB_DIR) run fmt:check
	npm --prefix $(WEB_DIR) run typecheck

fmt: ## autoformat source
	gofmt -w cmd internal
	if command -v goimports >/dev/null 2>&1; then goimports -w cmd internal; fi
	npm --prefix $(WEB_DIR) run fmt

pages-preview: ## serve docs/ locally as GitHub Pages would
	./scripts/pages-preview.sh

release: ## run release checks and tag VERSION
	$(MAKE) data
	$(MAKE) build
	$(MAKE) test
	$(MAKE) smoke
	git tag $(VERSION)

clean: ## remove local build scratch
	rm -rf tmp dist-data $(PAGES_DIR)/assets

hooks-pre-commit:
	$(MAKE) fmt
	$(MAKE) lint
	gitleaks protect --staged --redact

hooks-commit-msg:
	./scripts/validate-commit-msg.sh "$$COMMIT_MSG_FILE"

hooks-pre-push:
	$(MAKE) test
	$(MAKE) build
	$(MAKE) smoke

hooks-post-merge:
	npm --prefix $(WEB_DIR) install
	go mod tidy

hooks-post-checkout:
	npm --prefix $(WEB_DIR) install
	go mod tidy
