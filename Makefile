PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE_VERSION = $(shell cat version.txt)
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
# This is automatically detected from the k8s.io/api version in go.mod
ENVTEST_K8S_VERSION ?= $(shell go list -m -f "{{ .Version }}" k8s.io/api | awk -F'[v.]' '{printf "1.%d", $$3}')
# ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script
# This is automatically detected from the sigs.k8s.io/controller-runtime version in go.mod
ENVTEST_VERSION ?= $(shell go list -m -f "{{ .Version }}" sigs.k8s.io/controller-runtime | awk -F'[v.]' '{printf "release-%d.%d", $$2, $$3}')

DEVICESHIFU_CMDS := deviceshifu/cmdhttp deviceshifu/cmdmqtt deviceshifu/cmdopcua deviceshifu/cmdplc4x deviceshifu/cmdsocket
HTTPSTUB_CMDS := httpstub/powershellstub httpstub/sshstub
SHIFUCTL_CMD := shifuctl
TELEMETRYSERVICE_CMD := telemetryservice
CONTROLLER_CMD := pkg/k8s/crd
BUILD_TARGETS := $(DEVICESHIFU_CMDS) $(HTTPSTUB_CMDS) $(SHIFUCTL_CMD) $(TELEMETRYSERVICE_CMD)

fmt: ## Run go fmt against code.
	go fmt ./...

vet: ## Run go vet against code.
	go vet ./...

install: shifuctl
	go install ${PROJECT_ROOT}/cmd/shifuctl

shifuctl:
	cd ${PROJECT_ROOT}/cmd/shifuctl; go build

clean:
	rm -f ${PROJECT_ROOT}/cmd/shifuctl/shifuctl

build:
	for target in $(BUILD_TARGETS); do \
		go build -o bin/$$target  ./cmd/$$target; \
		echo "finished building $$target"; \
	done
	echo "building controller"
	cd $(CONTROLLER_CMD) && go build -o bin/shifu-controller main.go

.PHONY: test
test: fmt envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -v -race -coverprofile=coverage.out -covermode=atomic $(shell go list ./... | grep -v -E '/cmd|/mockdevice|/pkg/telemetryservice')
	go test -v -coverprofile=coverage_tdengine.out -covermode=atomic ./pkg/telemetryservice/...
	
buildx-push-image-deviceshifu-http-http:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuHTTP \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-http:${IMAGE_VERSION} --push

buildx-push-image-deviceshifu-http-mqtt:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuMQTT \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --push

buildx-push-image-deviceshifu-http-socket:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuSocket \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-socket:${IMAGE_VERSION} --push

buildx-push-image-deviceshifu-http-opcua:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuOPCUA \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-opcua:${IMAGE_VERSION} --push

buildx-push-image-deviceshifu-http-plc4x:
	docker buildx build --platform=linux/amd64,linux/arm64 -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuPLC4X \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-plc4x:${IMAGE_VERSION} --push

buildx-push-image-deviceshifu-tcp-tcp:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuTCP\
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-tcp-tcp:${IMAGE_VERSION} --push

buildx-push-image-deviceshifu-http-lwm2m:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm  -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuLwM2M \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-lwm2m:${IMAGE_VERSION} --push

buildx-push-image-gateway-lwm2m:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm \
		-f $(PROJECT_ROOT)/dockerfiles/Dockerfile.gatewayLwM2M \
		--build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
		-t edgehub/gateway-lwm2m:$(IMAGE_VERSION) --push

buildx-push-image-shifu-controller:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f $(PROJECT_ROOT)/pkg/k8s/crd/Dockerfile \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/shifu-controller:$(IMAGE_VERSION) --push

buildx-push-image-mockdevice-thermometer:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/thermometer/Dockerfile.mockdevice-thermometer \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-thermometer:$(IMAGE_VERSION) --push

buildx-push-image-mockdevice-robot-arm:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/robot-arm/Dockerfile.mockdevice-robot-arm \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-robot-arm:$(IMAGE_VERSION) --push

buildx-push-image-mockdevice-plate-reader:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/plate-reader/Dockerfile.mockdevice-plate-reader \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-plate-reader:$(IMAGE_VERSION) --push

buildx-push-image-mockdevice-agv:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/agv/Dockerfile.mockdevice-agv \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-agv:$(IMAGE_VERSION) --push

buildx-push-image-mockdevice-plc:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/plc/Dockerfile.mockdevice-plc \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-plc:$(IMAGE_VERSION) --push

buildx-push-image-mockdevice-socket:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/socket/Dockerfile.mockdevice-socket \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-socket:$(IMAGE_VERSION) --push

buildx-push-image-mockdevice-opcua:
	docker buildx build --platform=linux/amd64,linux/arm64 \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/opcua/Dockerfile.mockdevice-opcua \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-opcua:$(IMAGE_VERSION) --push

.PHONY: buildx-push-image-deviceshifu
buildx-push-image-deviceshifu: \
	buildx-push-image-deviceshifu-http-http \
	buildx-push-image-deviceshifu-http-mqtt \
	buildx-push-image-deviceshifu-http-socket \
	buildx-push-image-deviceshifu-http-opcua \
	buildx-push-image-deviceshifu-http-plc4x \
	buildx-push-image-deviceshifu-tcp-tcp \
	buildx-push-image-deviceshifu-http-lwm2m \
	buildx-push-image-gateway-lwm2m

buildx-push-image-telemetry-service:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.telemetryservice \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/telemetryservice:${IMAGE_VERSION} --push

buildx-build-image-shifu-controller:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f $(PROJECT_ROOT)/pkg/k8s/crd/Dockerfile \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/shifu-controller:$(IMAGE_VERSION) --load

buildx-build-image-deviceshifu-http-http:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuHTTP \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-http:${IMAGE_VERSION} --load

buildx-build-image-deviceshifu-http-lwm2m:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuLwM2M \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-lwm2m:${IMAGE_VERSION} --load

buildx-build-image-deviceshifu-http-mqtt:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuMQTT \
	 	--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --load

buildx-build-image-deviceshifu-http-socket:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuSocket \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-socket:${IMAGE_VERSION} --load

buildx-build-image-deviceshifu-http-opcua:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuOPCUA \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-opcua:${IMAGE_VERSION} --load

buildx-build-image-deviceshifu-http-plc4x:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuPLC4X\
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-plc4x:${IMAGE_VERSION} --load

buildx-build-image-deviceshifu-tcp-tcp:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuTCP\
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-tcp-tcp:${IMAGE_VERSION} --load

buildx-build-image-mockdevice-thermometer:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/thermometer/Dockerfile.mockdevice-thermometer \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-thermometer:$(IMAGE_VERSION) --load

buildx-build-image-mockdevice-robot-arm:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/robot-arm/Dockerfile.mockdevice-robot-arm \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-robot-arm:$(IMAGE_VERSION) --load

buildx-build-image-mockdevice-plate-reader:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/plate-reader/Dockerfile.mockdevice-plate-reader \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-plate-reader:$(IMAGE_VERSION) --load

buildx-build-image-mockdevice-agv:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/agv/Dockerfile.mockdevice-agv \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-agv:$(IMAGE_VERSION) --load

buildx-build-image-mockdevice-plc:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/plc/Dockerfile.mockdevice-plc \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-plc:$(IMAGE_VERSION) --load

buildx-build-image-mockdevice-socket:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/socket/Dockerfile.mockdevice-socket \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-socket:$(IMAGE_VERSION) --load

buildx-build-image-mockdevice-opcua:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/examples/deviceshifu/mockdevice/opcua/Dockerfile.mockdevice-opcua \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/mockdevice-opcua:$(IMAGE_VERSION) --load

buildx-build-image-deviceshifu: \
	buildx-build-image-deviceshifu-http-http \
	buildx-build-image-deviceshifu-http-mqtt \
	buildx-build-image-deviceshifu-http-socket \
	buildx-build-image-deviceshifu-http-opcua \
	buildx-build-image-deviceshifu-http-plc4x \
	buildx-build-image-deviceshifu-tcp-tcp \
	buildx-build-image-deviceshifu-http-lwm2m \
	buildx-build-image-gateway-lwm2m

buildx-build-image-telemetry-service:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.telemetryservice\
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/telemetryservice:${IMAGE_VERSION} --load

buildx-build-image-gateway-lwm2m:
	docker buildx build --platform=linux/$(shell go env GOARCH) \
          -f $(PROJECT_ROOT)/dockerfiles/Dockerfile.gatewayLwM2M \
          --build-arg PROJECT_ROOT="$(PROJECT_ROOT)" $(PROJECT_ROOT) \
          -t edgehub/gateway-lwm2m:$(IMAGE_VERSION) --load

.PHONY: download-demo-files
download-demo-files:
	docker pull edgehub/mockdevice-agv:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-plate-reader:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-robot-arm:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-thermometer:${IMAGE_VERSION}
	docker pull edgehub/deviceshifu-http-http:${IMAGE_VERSION}
	docker pull edgehub/shifu-controller:${IMAGE_VERSION}
	docker pull nginx:1.21

docker-push-image-deviceshifu:
	docker push edgehub/deviceshifu-http-http:${IMAGE_VERSION}

docker-push-deviceshifu-demo-image:
	docker push edgehub/demo-image-alpine:${IMAGE_VERSION}

docker-push-image-mockdevices:
	docker push edgehub/mockdevice-thermometer:${IMAGE_VERSION}
	docker push edgehub/mockdevice-robot-arm:${IMAGE_VERSION}
	docker push edgehub/mockdevice-plate-reader:${IMAGE_VERSION}
	docker push edgehub/mockdevice-agv:${IMAGE_VERSION}

.PHONY: clean-images
clean-images:
	docker rmi $(shell sudo docker images | grep 'edgehub')

tag:
	go run tools/tag.go ${PROJECT_ROOT} ${IMAGE_VERSION} $(VERSION)
	cd pkg/k8s/crd/ && (make generate-controller-yaml IMG=edgehub/shifu-controller:$(VERSION) generate-install-yaml)
	sed -e "s/${IMAGE_VERSION}/${VERSION}/g" ./test/scripts/deviceshifu-demo-aio.sh > ./test/scripts/tmp.sh && mv ./test/scripts/tmp.sh ./test/scripts/deviceshifu-demo-aio.sh
	echo $(VERSION) > version.txt

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest

## Tool Versions
KUSTOMIZE_VERSION ?= v5.7.1
CONTROLLER_TOOLS_VERSION ?= v0.19.0

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || { curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN); }

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download setup-envtest locally if necessary.
$(ENVTEST): $(LOCALBIN)
	$(call go-install-tool,$(ENVTEST),sigs.k8s.io/controller-runtime/tools/setup-envtest,$(ENVTEST_VERSION))

.PHONY: setup-envtest
setup-envtest: envtest ## Download the binaries required for ENVTEST in the local bin directory.
	@echo "Setting up envtest binaries for Kubernetes version $(ENVTEST_K8S_VERSION)..."
	@$(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path || { \
		echo "Error: Failed to set up envtest binaries for version $(ENVTEST_K8S_VERSION)."; \
		exit 1; \
	}

# go-install-tool will 'go install' any package with custom target and name of binary, if it doesn't exist
# $1 - target path with name of binary
# $2 - package url which can be installed
# $3 - specific version of package
define go-install-tool
@[ -f "$(1)-$(3)" ] && [ "$$(readlink -- "$(1)" 2>/dev/null)" = "$(1)-$(3)" ] || { \
set -e; \
package=$(2)@$(3) ;\
echo "Downloading $${package}" ;\
rm -f $(1) ;\
GOBIN=$(LOCALBIN) go install $${package} ;\
mv $(1) $(1)-$(3) ;\
} ;\
ln -sf $$(realpath $(1)-$(3)) $(1)
endef
