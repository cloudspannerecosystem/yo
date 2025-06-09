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

build: regen ## build yo command and regenerate template bin
	go build

regen: tplbin/templates.go ## regenerate template bin

deps:
	go get -u github.com/jessevdk/go-assets-builder

.PHONY: gomod
gomod: ## Run go mod
	go mod tidy
	cd v2; go mod tidy

.PHONY: lint
lint: ## Run linters
	go fmt ./...
	go vet ./...

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
	$(MAKE) -j4 testdata/default testdata/underscore testdata/customtypes testdata/single

testdata/default:
	rm -rf test/testmodels/default && mkdir -p test/testmodels/default
	$(YOBIN) $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --package models --out test/testmodels/default/

testdata/underscore:
	rm -rf test/testmodels/underscore && mkdir -p test/testmodels/underscore
	$(YOBIN) $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --package models --underscore --out test/testmodels/underscore/

testdata/single:
	rm -rf test/testmodels/single && mkdir -p test/testmodels/single
	$(YOBIN) $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --out test/testmodels/single/single_file.go --single-file

testdata/customtypes:
	rm -rf test/testmodels/customtypes && mkdir -p test/testmodels/customtypes
	$(YOBIN) $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --custom-types-file test/testdata/custom_column_types.yml --out test/testmodels/customtypes/

testdata-from-ddl:
	$(MAKE) -j4 testdata-from-ddl/default testdata-from-ddl/underscore testdata-from-ddl/customtypes testdata-from-ddl/single

testdata-from-ddl/default:
	rm -rf test/testmodels/default && mkdir -p test/testmodels/default
	$(YOBIN) generate ./test/testdata/schema.sql --from-ddl --package models --out test/testmodels/default/

testdata-from-ddl/underscore:
	rm -rf test/testmodels/underscore && mkdir -p test/testmodels/underscore
	$(YOBIN) generate ./test/testdata/schema.sql --from-ddl --package models --underscore --out test/testmodels/underscore/

testdata-from-ddl/single:
	rm -rf test/testmodels/single && mkdir -p test/testmodels/single
	$(YOBIN) generate ./test/testdata/schema.sql --from-ddl --out test/testmodels/single/single_file.go --single-file

testdata-from-ddl/customtypes:
	rm -rf test/testmodels/customtypes && mkdir -p test/testmodels/customtypes
	$(YOBIN) generate ./test/testdata/schema.sql --from-ddl --custom-types-file test/testdata/custom_column_types.yml --out test/testmodels/customtypes/

recreate-templates:: ## recreate templates
	rm -rf templates && mkdir templates
	$(YOBIN) create-template --template-path templates

.PHONY: check_lint
check_lint: lint ## check linter errors
	if git diff --quiet; then \
        exit 0; \
	else \
		echo "\nerror: make lint resulted in a change of files."; \
		echo "Please run make lint locally before pushing."; \
		exit 1; \
	fi

.PHONY: check_gomod
check_gomod: gomod ## check whether or not go mod tidy has been run
	if git diff --quiet go.mod go.sum v2/go.mod v2/go.sum; then \
        exit 0; \
	else \
		echo "\nerror: make gomod resulted in a change of files."; \
		echo "Please run make gomod locally before pushing."; \
		exit 1; \
	fi

.PHONY: check-diff

EXPECTED_FILES := \
	test/testmodels/customtypes/compositeprimarykey.yo.go \
	test/testmodels/default/compositeprimarykey.yo.go \
	test/testmodels/single/single_file.go \
	test/testmodels/underscore/composite_primary_key.yo.go

check-diff:
	@echo "Checking git diff against expected files..."
	@ACTUAL_FILES=$$(git diff --name-only | sort) ; \
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