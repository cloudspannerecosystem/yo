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

regen: module/builtin/tplbin/templates.go ## regenerate template bin

deps:
	go get -u github.com/jessevdk/go-assets-builder

module/builtin/tplbin/templates.go: $(wildcard module/builtin/templates/*.tpl)
	mkdir -p module/builtin/tplbin
	go-assets-builder \
		--package=tplbin \
		--strip-prefix="/module/builtin/templates/" \
		--output module/builtin/tplbin/templates.go \
		module/builtin/templates/*.tpl

.PHONY: test
test: ## run test
	@echo run tests with spanner emulator
	go test -race -v ./...

recreate-templates:: ## recreate templates
	rm -rf module/builtin/templates && mkdir module/builtin/templates
	$(YOBIN) create-template --template-path module/builtin/templates

USE_DDL ?= false
ifeq ($(USE_DDL),true)
GENERATE_OPT = ./test/testdata/schema.sql --from-ddl
else
GENERATE_OPT = $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME)
endif

testdata: ## generate test models
	$(MAKE) -j4 testdata/default testdata/legacy_default testdata/customtypes testdata/dump_types

testdata-from-ddl: ## generate test models
	$(MAKE) USE_DDL=true -j4 testdata/default testdata/customtypes testdata/dump_types

testdata/default:
	rm -rf test/testmodels/default && mkdir -p test/testmodels/default
	$(YOBIN) generate $(GENERATE_OPT) --package models --out test/testmodels/default/

testdata/legacy_default:
	rm -rf test/testmodels/legacy_default && mkdir -p test/testmodels/legacy_default
	$(YOBIN) generate $(GENERATE_OPT) --use-legacy-index-module --package models --out test/testmodels/legacy_default/

testdata/customtypes:
	rm -rf test/testmodels/customtypes && mkdir -p test/testmodels/customtypes
	$(YOBIN) generate $(GENERATE_OPT) --config test/testdata/config.yml --out test/testmodels/customtypes/

testdata/dump_types:
	rm -rf test/testmodels/dump_types && mkdir -p test/testmodels/dump_types
	$(YOBIN) generate $(GENERATE_OPT) --suffix '.txt' --disable-format --disable-default-modules --type-module test/testdata/dump_types.go.tpl --package models --out test/testmodels/dump_types/
