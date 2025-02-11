TAG := $(shell git rev-parse --short HEAD)
DIR := $(shell pwd -L)

dep:
	docker run -i \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        asecurityteam/sdcli:v1 go dep

lint:
	docker run -i \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        asecurityteam/sdcli:v1 go lint

test:
	docker run -i \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        asecurityteam/sdcli:v1 go test

integration:
	docker run -i \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        asecurityteam/sdcli:v1 go integration

coverage:
	docker run -i \
        --mount src="$(DIR)",target="$(DIR)",type="bind" \
        -w "$(DIR)" \
        asecurityteam/sdcli:v1 go coverage

doc: ;

build-dev: ;

build: ;

run: ;

deploy-dev: ;

deploy: ;
