export SPANNER_PROJECT_NAME  ?= mercari-example-project
export SPANNER_INSTANCE_NAME ?= mercari-example-instance
export SPANNER_DATABASE_NAME ?= yo-test

YOBIN ?= yo

all: build

build: regen
	go build

regen: tplbin/templates.go

deps:
	go get -u github.com/golang/dep/cmd/dep
	go get -u github.com/jessevdk/go-assets-builder
	go get -u golang.org/x/tools/cmd/goimports

tplbin/templates.go: $(wildcard templates/*.tpl)
	go-assets-builder \
		--package=tplbin \
		--strip-prefix="/templates/" \
		--output tplbin/templates.go \
		templates/*.tpl

.PHONY: test
test:
	go test -race -v ./test

testsetup:
	@gcloud --project $(SPANNER_PROJECT_NAME) spanner databases create $(SPANNER_DATABASE_NAME) --instance $(SPANNER_INSTANCE_NAME) --ddl "$(shell cat ./test/testdata/schema.sql)"

testdata:
	$(MAKE) -j4 testdata/default testdata/customtypes testdata/single

testdata/default:
	rm -rf test/testmodels/default && mkdir -p test/testmodels/default
	$(YOBIN) $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --package models --out test/testmodels/default/

testdata/single:
	rm -rf test/testmodels/single && mkdir -p test/testmodels/single
	$(YOBIN) $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --out test/testmodels/single/single_file.go --single-file

testdata/customtypes:
	rm -rf test/testmodels/customtypes && mkdir -p test/testmodels/customtypes
	$(YOBIN) $(SPANNER_PROJECT_NAME) $(SPANNER_INSTANCE_NAME) $(SPANNER_DATABASE_NAME) --custom-types-file test/testdata/custom_column_types.yml --out test/testmodels/customtypes/

recreate-templates::
	rm -rf templates && mkdir templates
	$(YOBIN) --create-templates --template-path templates
