
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

tidy:
	cd log && go mod tidy;
	cd internal && go mod tidy;
	cd testlogger && go mod tidy;
	cd middleware && go mod tidy;
	cd logrus && go mod tidy;
	cd zerolog && go mod tidy;
	go mod tidy
