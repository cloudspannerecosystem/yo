export SPANNER_EMULATOR_HOST ?= localhost:9010
export SPANNER_EMULATOR_HOST_REST ?= localhost:9020
export SPANNER_PROJECT_NAME ?= yo-test
export SPANNER_INSTANCE_NAME ?= yo-test
export SPANNER_DATABASE_NAME ?= yo-test

YOBIN ?= yo

export GO111MODULE=on

.PHONY: help
help: ## show this help message.
	@grep -hE '^\S+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: build

build: regen ## build yo command and regenerate template bin
	go build

regen: tplbin/templates.go ## regenerate template bin

deps:
	go get -u github.com/jessevdk/go-assets-builder

tplbin/templates.go: $(wildcard templates/*.tpl)
	go-assets-builder \
		--package=tplbin \
		--strip-prefix="/templates/" \
		--output tplbin/templates.go \
		templates/*.tpl

.PHONY: test
test: ## run test
	@echo run tests with spanner emulator
	go test -race -v ./test

testdata: ## generate test models
	$(MAKE) -j4 testdata/default testdata/customtypes

testdata/default:
	rm -rf test/testmodels/default && mkdir -p test/testmodels/default
	$(YOBIN) generate $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --package models --out test/testmodels/default/

testdata/customtypes:
	rm -rf test/testmodels/customtypes && mkdir -p test/testmodels/customtypes
	$(YOBIN) generate $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --custom-types-file test/testdata/custom_column_types.yml --out test/testmodels/customtypes/

testdata-from-ddl:
	$(MAKE) -j4 testdata-from-ddl/default testdata-from-ddl/customtypes

testdata-from-ddl/default:
	rm -rf test/testmodels/default && mkdir -p test/testmodels/default
	$(YOBIN) generate ./test/testdata/schema.sql --from-ddl --package models --out test/testmodels/default/

testdata-from-ddl/customtypes:
	rm -rf test/testmodels/customtypes && mkdir -p test/testmodels/customtypes
	$(YOBIN) generate ./test/testdata/schema.sql --from-ddl --custom-types-file test/testdata/custom_column_types.yml --out test/testmodels/customtypes/

recreate-templates:: ## recreate templates
	rm -rf templates && mkdir templates
	$(YOBIN) create-template --template-path templates
