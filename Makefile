
MODULES := $(shell find . -type f -name go.mod -print0 | xargs -0 dirname)

.DEFAULT_GOAL := help
.PHONY: help lint test

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: $(patsubst %,%.lint,$(MODULES)) ## Runs go vet and golangci-lint.
test: $(patsubst %,%.test,$(MODULES)) ## Run tests.

%.lint:
	cd $* && go vet ./... && golangci-lint run ;

%.test:
	cd $* && go test -short -v ./... -race;

tidy: ## Run go mod tidy on all modules in the repo.
	find . -type f -name go.mod -execdir go mod tidy -e ';'
