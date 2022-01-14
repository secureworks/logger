PACKAGES = $(shell go list ./...)

.DEFAULT_GOAL := help
.PHONY: help lint test

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: ## Run go vet and golangci-lint.
	go vet ./... || go clean ./...; go vet ./... && golangci-lint run ./...

flake: ## Run tests.
	go test -short -v ./... -test.failfast -test.count 10 -race
