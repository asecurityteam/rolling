.PHONY: docker-build-go docker-build-lint docker-build dep lint coverage test

TAG := $(shell git rev-parse --short HEAD)
DIR := $(shell pwd -L)
LOCAL_GO_IMAGE ?= rolling-go
LOCAL_LINT_IMAGE ?= rolling-golangci-lint
GODOCKER = docker run --rm -v "$(DIR):$(DIR)" -w "$(DIR)" $(LOCAL_GO_IMAGE)
LINTDOCKER = docker run --rm -v "$(DIR):$(DIR)" -w "$(DIR)" $(LOCAL_LINT_IMAGE)

COVERAGE_DIR := .coverage
UNIT_COVERAGE_DIR := $(COVERAGE_DIR)/unit
UNIT_COVERAGE_FILE := $(UNIT_COVERAGE_DIR)/unit.cover.out

docker-build-go:
	docker build --target go -t $(LOCAL_GO_IMAGE) .
 
docker-build-lint:
	docker build --target lint -t $(LOCAL_LINT_IMAGE) -f linter.Dockerfile .

docker-build: docker-build-go docker-build-lint
 
dep: docker-build-go
	$(GODOCKER) go mod vendor

lint: docker-build-lint
	$(LINTDOCKER) golangci-lint run --config .golangci.yaml ./... -v

coverage-setup:
	mkdir -p $(UNIT_COVERAGE_DIR)
	touch $(UNIT_COVERAGE_FILE)

test: coverage-setup docker-build-go
	$(GODOCKER) go test -coverprofile=$(UNIT_COVERAGE_FILE) -v -race ./...

integration: ;
 
coverage: docker-build-go
	$(GODOCKER) go tool cover -func=$(UNIT_COVERAGE_FILE)

doc: ;

build-dev: ;

build: ;

run: ;

deploy-dev: ;

deploy: ;
