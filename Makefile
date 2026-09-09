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
## install: install the binary and its man page; PREFIX=/usr/local for a system path
install: build
ifeq ($(strip $(PREFIX)),)
	@mkdir -p "$(BINDIR)"
	@cp $(BIN) "$(BINDIR)/$(BIN)"
	install -d "$(USER_MANDIR)"
	install -m 0644 man/$(BIN).1 "$(USER_MANDIR)/$(BIN).1"
	@echo "installed $(BINDIR)/$(BIN) and $(USER_MANDIR)/$(BIN).1"
else
	install -d "$(DESTDIR)$(PREFIX)/bin" "$(DESTDIR)$(PREFIX)/share/man/man1"
	install -m 0755 $(BIN) "$(DESTDIR)$(PREFIX)/bin/$(BIN)"
	install -m 0644 man/$(BIN).1 "$(DESTDIR)$(PREFIX)/share/man/man1/$(BIN).1"
	@echo "installed $(DESTDIR)$(PREFIX)/bin/$(BIN)"
endif

.PHONY: uninstall
## uninstall: remove what `make install` put down
uninstall:
ifeq ($(strip $(PREFIX)),)
	rm -f "$(BINDIR)/$(BIN)" "$(USER_MANDIR)/$(BIN).1"
else
	rm -f "$(DESTDIR)$(PREFIX)/bin/$(BIN)" \
		"$(DESTDIR)$(PREFIX)/share/man/man1/$(BIN).1"
endif

##@ Quality

.PHONY: lint
lint: ## Run all pre-commit checks on the whole tree
	pre-commit run --all-files

.PHONY: test
test: ## Run the test suite
	go test ./...

.PHONY: run
run: ## Run the binary
	go run . $(ARGS)

.PHONY: format
format: ## Rewrite the sources to gofmt form
	gofmt -w .

.PHONY: analyze
analyze: ## Lint with the house rule set
	golangci-lint run ./...
