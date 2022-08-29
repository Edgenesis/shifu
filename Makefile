PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE_VERSION = $(shell cat version.txt)

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

.PHONY: buildx-push-image-deviceshifu
buildx-push-image-deviceshifu: \
	buildx-push-image-deviceshifu-http-http \
	buildx-push-image-deviceshifu-http-mqtt \
	buildx-push-image-deviceshifu-http-socket \
	buildx-push-image-deviceshifu-http-opcua

buildx-build-image-deviceshifu-http-http:
	docker buildx build --platform=linux/$(shell go env GOARCH) -f ${PROJECT_ROOT}/dockerfiles/Dockerfile.deviceshifuHTTP \
		--build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} \
		-t edgehub/deviceshifu-http-http:${IMAGE_VERSION} --load

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

buildx-build-image-deviceshifu: \
	buildx-build-image-deviceshifu-http-http \
	buildx-build-image-deviceshifu-http-mqtt \
	buildx-build-image-deviceshifu-http-socket \
	buildx-build-image-deviceshifu-http-opcua

.PHONY: download-demo-files
download-demo-files:
	docker pull edgehub/mockdevice-agv:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-plate-reader:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-robot-arm:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-thermometer:${IMAGE_VERSION}
	docker pull edgehub/deviceshifu-http-http:${IMAGE_VERSION}
	docker pull edgehub/shifu-controller:${IMAGE_VERSION}
	docker pull quay.io/brancz/kube-rbac-proxy:v0.12.0
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
