TAG := $(shell git rev-parse --short HEAD)
DIR := $(shell pwd -L)
GOPATH := ${GOPATH}
ifeq ($(GOPATH),)
	GOPATH := ${HOME}/go
endif
PROJECT_PATH := $(subst $(GOPATH)/src/,,$(DIR))

dep:
	docker run -ti \
        --mount src="$(DIR)",target="/go/src/$(PROJECT_PATH)",type="bind" \
        -w "/go/src/$(PROJECT_PATH)" \
        asecurityteam/sdcli:v1 go dep

lint:
	docker run -ti \
        --mount src="$(DIR)",target="/go/src/$(PROJECT_PATH)",type="bind" \
        -w "/go/src/$(PROJECT_PATH)" \
        asecurityteam/sdcli:v1 go lint

test:
	docker run -ti \
        --mount src="$(DIR)",target="/go/src/$(PROJECT_PATH)",type="bind" \
        -w "/go/src/$(PROJECT_PATH)" \
        asecurityteam/sdcli:v1 go test

integration:
	docker run -ti \
        --mount src="$(DIR)",target="/go/src/$(PROJECT_PATH)",type="bind" \
        -w "/go/src/$(PROJECT_PATH)" \
        asecurityteam/sdcli:v1 go integration

coverage:
	docker run -ti \
        --mount src="$(DIR)",target="/go/src/$(PROJECT_PATH)",type="bind" \
        -w "/go/src/$(PROJECT_PATH)" \
        asecurityteam/sdcli:v1 go coverage

doc: ;

build-dev: ;

build: ;

run: ;

deploy-dev: ;

deploy: ;
