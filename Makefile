MODULES := $(shell go list -f '{{.Dir}}' -m)

.DEFAULT_GOAL := help
.PHONY: help lint test tidy

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: $(patsubst %,%.lint,$(MODULES)) ## Run go vet and golangci-lint.

%.lint:
	cd $* && go vet ./... && golangci-lint run ;

test: ## Run tests.
	@for mod in $(MODULES); do \
	  printf "\nRunning tests for module at $${mod}/go.mod:\n"; \
	  go test -short -v -C "$${mod}" ./... -race ; \
	done

tidy: ## Run go mod tidy.
	go mod tidy -v -C ./log;
	go mod tidy -v -C ./internal;
	go mod tidy -v -C ./testlogger;
	go mod tidy -v -C ./middleware;
	go mod tidy -v -C ./logrus;
	go mod tidy -v -C ./zerolog;
	go mod tidy -v -C ./logrus;
	go mod tidy
