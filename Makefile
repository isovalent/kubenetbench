SHELL=/bin/bash

.PHONY: docker-images

all: kubenetbench/kubenetbench

DOCKER_USER ?= kkourt
REPO=github.com/kkourt/kubenetbench


kubenetbench/kubenetbench: FORCE
	pushd $(CURDIR)/kubenetbench && go build && popd

docker-images:
	docker build . -f Dockerfile.kubenetbench -t $(DOCKER_USER)/kubenetbench
	docker push $(DOCKER_USER)/kubenetbench

FORCE:
