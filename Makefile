PACKAGES = $(shell go list ./...)

.DEFAULT_GOAL := help
.PHONY: help lint test

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-10s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

lint: ## Run go vet and golangci-lint.
	go vet ./... || go clean ./...; go vet ./... && golangci-lint run ./...

test: ## Run tests.
	go test -short -v ./... -race;
	cd middleware && go test -short -v ./... -race;
	cd logrus && go test -short -v ./... -race;
	cd zerolog && go test -short -v ./... -race;
