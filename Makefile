.PHONY: help setup lint test build install

BINDIR ?= $(HOME)/.local/bin

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN {FS = ":.*?## "}; {printf "  %-10s %s\n", $$1, $$2}'

setup: ## Install the pre-commit hook
	pre-commit install

lint: ## Run all pre-commit checks on the whole tree
	pre-commit run --all-files

test: ## Run the test suite
	go test ./...

build: ## Build the offline binary
	go build -o offline offline.go

install: build ## Build and drop the binary in BINDIR (default ~/.local/bin)
	@mkdir -p "$(BINDIR)"
	@cp offline "$(BINDIR)/offline"
	@echo "installed $(BINDIR)/offline"
