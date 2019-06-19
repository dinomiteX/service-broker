# If the USE_SUDO_FOR_DOCKER env var is set, prefix docker commands with 'sudo'
ifdef USE_SUDO_FOR_DOCKER
	SUDO_CMD = sudo
endif

IMAGE ?= dinomitex/service-broker
TAG ?= latest
PULL ?= Always

build: ## Builds the starter pack
	go build -i github.com/dinomiteX/service-broker/cmd/servicebroker

test: ## Runs the tests
	go test -v $(shell go list ./... | grep -v /vendor/ | grep -v /test/)

linux: ## Builds a Linux executable
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
	go build -o servicebroker-linux --ldflags="-s" github.com/dinomiteX/service-broker/cmd/servicebroker

image: linux ## Builds a Linux based image
	cp servicebroker-linux image/servicebroker
	$(SUDO_CMD) docker build image/ -t "$(IMAGE):$(TAG)"

clean: ## Cleans up build artifacts
	rm -f servicebroker
	rm -f servicebroker-linux
	rm -f image/servicebroker

push: image ## Pushes the image to dockerhub, REQUIRES SPECIAL PERMISSION
	$(SUDO_CMD) docker push "$(IMAGE):$(TAG)"

deploy-helm: ## Deploys image with helm
	helm upgrade --install service-broker-dino --namespace service-broker-dino \
	charts/servicebroker \
	--set image="$(IMAGE):$(TAG)",imagePullPolicy="$(PULL)"
	
create-ns: ## Cleans up the namespaces
	kubectl create ns dino-instance || true

provision: create-ns ## Provisions a service instance
	kubectl apply -f manifests/service-instance.yaml

bind: ## Creates a binding
	kubectl apply -f manifests/service-binding.yaml

delns: ## Deletes Namespace, Service Instance , Service Binding
	kubectl delete ns service-broker-dino
	kubectl delete ns dino-instance

pods: ## Get Service Broker Pod
	kubectl get po -n service-broker-dino

help: ## Shows the help
	@echo 'Usage: make <OPTIONS> ... <TARGETS>'
	@echo ''
	@echo 'Available targets are:'
	@echo ''
	@grep -E '^[ a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
        awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ''

.PHONY: ibuild test linux image clean push deploy-helm deploy-openshift create-ns provision bind delns pods help
