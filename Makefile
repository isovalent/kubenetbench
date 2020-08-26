.PHONY: docker-images

all: kubenetbench/kubenetbench benchmonitor/srv/srv

DOCKER_USER ?= kkourt
REPO=github.com/kkourt/kubenetbench

GO ?= go

kubenetbench/kubenetbench: FORCE
	cd $(CURDIR)/kubenetbench && $(GO) build

benchmonitor/api/benchmonitor.pb.go: benchmonitor/benchmonitor.proto
	protoc  $< --go_out=plugins=grpc:benchmonitor

benchmonitor/srv/srv: FORCE benchmonitor/api/benchmonitor.pb.go
	cd $(CURDIR)/benchmonitor/srv && $(GO) build

docker-images:
	docker build . -f Dockerfile.knb -t $(DOCKER_USER)/kubenetbench
	docker push $(DOCKER_USER)/kubenetbench
	#
	docker build -f Dockerfile.knb-monitor . -t $(DOCKER_USER)/kubenetbench-monitor
	docker push $(DOCKER_USER)/kubenetbench-monitor


FORCE:
