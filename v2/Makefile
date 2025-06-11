export SPANNER_EMULATOR_HOST ?= localhost:9010
export SPANNER_EMULATOR_HOST_REST ?= localhost:9020
export SPANNER_PROJECT_NAME ?= yo-test
export SPANNER_INSTANCE_NAME ?= yo-test
export SPANNER_DATABASE_NAME ?= yo-test

YOBIN ?= yo

.PHONY: help
help: ## show this help message.
	@grep -hE '^\S+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

all: build

build: ## build yo command and regenerate template bin
	go build

deps:
	go mod download

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
	$(MAKE) -j4 testdata/default testdata/legacy_default testdata/dump_types

testdata-from-ddl: ## generate test models
	$(MAKE) USE_DDL=true testdata

testdata/default:
	rm -rf test/testmodels/default && mkdir -p test/testmodels/default
	$(YOBIN) generate $(GENERATE_OPT) --config test/testdata/config.yml --package models --out test/testmodels/default/

testdata/legacy_default:
	rm -rf test/testmodels/legacy_default && mkdir -p test/testmodels/legacy_default
	$(YOBIN) generate $(GENERATE_OPT) --config test/testdata/config.yml --use-legacy-index-module --package models --out test/testmodels/legacy_default/

testdata/dump_types:
	rm -rf test/testmodels/dump_types && mkdir -p test/testmodels/dump_types
	$(YOBIN) generate $(GENERATE_OPT) --suffix '.txt' --disable-format --disable-default-modules --type-module test/testdata/dump_types.go.tpl --package models --out test/testmodels/dump_types/

.PHONY: check-diff

EXPECTED_FILES := \
	v2/test/testmodels/default/composite_primary_key.yo.go

check-diff:
	@echo "Checking git diff against expected files..."
	@ACTUAL_FILES=$$(git diff --name-only | grep -v '^go\.mod$$' | grep -v '^go\.sum$$' | sort) ; \
	SORTED_EXPECTED_FILES=$$(echo "$(EXPECTED_FILES)" | tr ' ' '\n' | sort) ; \
	if [ "$$ACTUAL_FILES" = "$$SORTED_EXPECTED_FILES" ]; then \
		echo "Success: git diff output matches the expected file list." ; \
	else \
		echo "Error: git diff output does not match the expected file list." ; \
		echo "--- Expected Files ---" ; \
		echo "$$SORTED_EXPECTED_FILES" ; \
		echo "--- Actual Files ---" ; \
		echo "$$ACTUAL_FILES" ; \
		exit 1 ; \
	fi