# offline — run a program with its network access fully isolated.
#
# Every verb this repo exposes lives here; `make` on its own prints them,
# grouped, straight out of the `##` comments below. Nothing about building,
# testing or installing should need knowledge that is not in this file.

BIN    := offline
BINDIR ?= $(HOME)/.local/bin

.DEFAULT_GOAL := help
# help is pure output; the recipe echo would only be noise.
.SILENT: help

##@ General

.PHONY: help
help: ## Show this help
	awk 'BEGIN {FS = ":.*## "} \
	  /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } \
	  /^[a-zA-Z_0-9-]+:.*## / { printf "  \033[36m%-10s\033[0m %s\n", $$1, $$2 }' \
	  $(MAKEFILE_LIST)

.PHONY: setup
setup: ## Install the pre-commit hook
	pre-commit install

##@ Build

.PHONY: build
build: ## Build the offline binary
	go build -o $(BIN) offline.go

.PHONY: install
install: build ## Build and drop the binary in BINDIR (default ~/.local/bin)
	@mkdir -p "$(BINDIR)"
	@cp $(BIN) "$(BINDIR)/$(BIN)"
	@echo "installed $(BINDIR)/$(BIN)"

##@ Quality

.PHONY: lint
lint: ## Run all pre-commit checks on the whole tree
	pre-commit run --all-files

.PHONY: test
test: ## Run the test suite
	go test ./...
