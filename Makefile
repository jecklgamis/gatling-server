IMAGE_NAME:=jecklgamis/gatling-server
IMAGE_TAG:=$(shell git rev-parse HEAD)

BUILD_BRANCH:=$(shell git rev-parse --abbrev-ref HEAD)
BUILD_VERSION:=$(shell git rev-parse HEAD)
BUILD_OS:=darwin
BUILD_ARCH:=amd64

ifeq ($(shell uname -s), Linux)
	BUILD_OS:=linux
	BUILD_ARCH:=amd64
endif

default: help
help:
	@echo Makefile targets
	@echo make dist - build server binaries
	@echo make image - build docker image
	@echo make run - run Docker image
	@echo make test - run short duration tests
	@echo make test-all - run all tests
	@echo make test-coverage - run tests with coverage
	@echo make clean - delete built artifacts
	@echo make release - release distribution
dist: clean test-coverage server-binaries ssl-certs
release:
	@echo "Check 1. Have you updated scripts/release-version file?"
	@echo "Check 2. Is make dist successful?"
	@read -p "Break now if any of your answers is NO. Otherwise, press <Enter> to continue"
	@rm -rf dist
	@$(CURDIR)/scripts/create-tag.sh
	@TARGET_OS=linux TARGET_ARCH=amd64 scripts/create-dist.sh
	@TARGET_OS=darwin TARGET_ARCH=amd64 scripts/create-dist.sh
	@$(CURDIR)/scripts/create-relnotes.sh
up: dist image run
image:
	@docker build -t $(IMAGE_NAME):$(IMAGE_TAG) -t $(IMAGE_NAME):latest .
run:
	docker run --memory 1g --cpus 1.5 -p 58080:58080 -p 58443:8443   -i -t $(IMAGE_NAME):$(IMAGE_TAG)
run-bash:
	@docker run -i -t $(IMAGE_NAME):$(IMAGE_TAG) /bin/bash
login:
	@docker exec -it `docker ps | grep $(IMAGE_NAME) | awk '{print $$1}'` /bin/bash
install-deps:
	@$(CURDIR)/scripts/download-gatling-distribution.sh
	@go get -u golang.org/x/lint/golint
LD_FLAGS:="-X github.com/jecklgamis/gatling-server/pkg/version.BuildVersion=$(BUILD_VERSION) \
		  -X github.com/jecklgamis/gatling-server/pkg/version.BuildBranch=$(BUILD_BRANCH)"

server-binaries: gatling-server-$(BUILD_OS)-$(BUILD_ARCH)  gatling-server-linux-amd64
gatling-server-$(BUILD_OS)-$(BUILD_ARCH):
	@echo "Building $@"
	@go build -ldflags $(LD_FLAGS) -o bin/gatling-server-$(BUILD_OS)-$(BUILD_ARCH) cmd/server/gatling-server.go
	@chmod +x bin/gatling-server*
gatling-server-linux-amd64:
	@echo "Building $@"
	@GOOS=linux GOARCH=amd64 go build -ldflags $(LD_FLAGS) -o bin/gatling-server-linux-amd64 cmd/server/gatling-server.go
	@chmod +x $(CURDIR)/bin/gatling-server-linux-amd64
clean:
	@rm -rf dist
	@rm -f $(CURDIR)/bin/*
	@go clean -testcache
	@rm -f server.crt
	@rm -f server.key
ssl-certs:
	@$(CURDIR)/scripts/generate-ssl-certs.sh
.PHONY: test
test:
	@echo Running tests
	@go test -short ./...
test-all:
	@echo Running all tests
	@echo Requires env : AWS_REGION=some-aws-region
	@echo Requires env : GATLING_SERVER_INCOMING_S3_URL=some-s3-url
	@echo Requires env : GATLING_SERVER_RESULTS_S3_URL=some-s3-url
	@go test ./...
test-coverage:
	@echo Running tests with coverage
	@go test -short -cover ./...
rebuilder:
	@$(CURDIR)/scripts/rebuilder/rebuilder.sh
lint:
	@$(CURDIR)/scripts/linter.sh
push:
	@echo "$(IMAGE_TAG)}"
	docker image push $(IMAGE_NAME):$(IMAGE_TAG)
	docker image push $(IMAGE_NAME):latest

