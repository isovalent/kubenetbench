
.PHONY: docker-images

all: kubenetbench

DOCKER_USER ?= kkourt


kubenetbench: FORCE
	go build

docker-images:
	docker build . -f Dockerfile.kubenetbench -t $(DOCKER_USER)/kubenetbench
	docker push $(DOCKER_USER)/kubenetbench


FORCE:
