PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE_VERSION = v0.0.1

.PHONY: build-image-deviceshifu
build-image-deviceshifu:
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-http:${IMAGE_VERSION}

.PHONY: buildx-push-image-deviceshifu
buildx-push-image-deviceshifu:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-http:${IMAGE_VERSION} --push
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifuMQTT --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --push
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifuSocket --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-socket:${IMAGE_VERSION} --push

buildx-load-image-deviceshifu:
	docker buildx build --platform=linux/amd64 -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-http:${IMAGE_VERSION} --load
	docker buildx build --platform=linux/amd64 -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifuMQTT --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-mqtt:${IMAGE_VERSION} --load
	docker buildx build --platform=linux/amd64 -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifuSocket --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-socket:${IMAGE_VERSION} --load	

.PHONY: download-demo-files
download-demo-files:
	docker pull edgehub/mockdevice-agv:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-plate-reader:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-robot-arm:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-thermometer:${IMAGE_VERSION}
	docker pull edgehub/deviceshifu-http-http:${IMAGE_VERSION}
	docker pull edgehub/edgedevice-controller-multi:${IMAGE_VERSION}
	docker pull quay.io/brancz/kube-rbac-proxy:v0.8.0
	docker pull kindest/node:v1.23.4@sha256:0e34f0d0fd448aa2f2819cfd74e99fe5793a6e4938b328f657c8e3f81ee0dfb9
	docker pull nginx:1.21

compress-demo-files:
	mkdir -p build_dir
	docker save quay.io/brancz/kube-rbac-proxy:v0.8.0 | gzip > build_dir/kube-rbac-proxy.tar.gz
	docker save edgehub/mockdevice-agv:${IMAGE_VERSION} | gzip > build_dir/mockdevice-agv.tar.gz
	docker save edgehub/mockdevice-plate-reader:${IMAGE_VERSION} | gzip > build_dir/mockdevice-plate-reader.tar.gz
	docker save edgehub/mockdevice-robot-arm:${IMAGE_VERSION} | gzip > build_dir/mockdevice-robot-arm.tar.gz
	docker save edgehub/mockdevice-thermometer:${IMAGE_VERSION} | gzip > build_dir/mockdevice-thermometer.tar.gz
	docker save edgehub/deviceshifu-http-http:${IMAGE_VERSION} | gzip > build_dir/deviceshifu-http-http.tar.gz
	docker save edgehub/edgedevice-controller-multi:${IMAGE_VERSION} | gzip > build_dir/edgedevice-controller-multi.tar.gz
	docker save kindest/node:v1.23.4@sha256:0e34f0d0fd448aa2f2819cfd74e99fe5793a6e4938b328f657c8e3f81ee0dfb9 | gzip > build_dir/kind-image.tar.gz
	docker save nginx:1.21 | gzip > build_dir/nginx.tar.gz

compress-edgenesis-files:
	mkdir -p build_dir
	docker save edgehub/mockdevice-agv:${IMAGE_VERSION} | gzip > build_dir/mockdevice-agv.tar.gz
	docker save edgehub/mockdevice-plate-reader:${IMAGE_VERSION} | gzip > build_dir/mockdevice-plate-reader.tar.gz
	docker save edgehub/mockdevice-robot-arm:${IMAGE_VERSION} | gzip > build_dir/mockdevice-robot-arm.tar.gz
	docker save edgehub/mockdevice-thermometer:${IMAGE_VERSION} | gzip > build_dir/mockdevice-thermometer.tar.gz
	docker save edgehub/deviceshifu-http-http:${IMAGE_VERSION} | gzip > build_dir/deviceshifu-http-http.tar.gz
	docker save edgehub/edgedevice-controller:${IMAGE_VERSION} | gzip > build_dir/edgedevice-controller.tar.gz

.PHONY: build-deviceshifu-demo-image
build-deviceshifu-demo-image:
	docker build -f ${PROJECT_ROOT}/Dockerfile.demo --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/demo-image-alpine:${IMAGE_VERSION}

buildx-load-deviceshifu-demo-image:
	docker buildx build --platform=linux/amd64 -f ${PROJECT_ROOT}/Dockerfile.demo --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/demo-image-alpine-multi:${IMAGE_VERSION} --load

buildx-push-deviceshifu-demo-image:
	docker buildx build --platform=linux/amd64,linux/arm64,darwin/arm64 -f ${PROJECT_ROOT}/Dockerfile.demo --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/demo-image-alpine-multi:${IMAGE_VERSION} --push

.PHONY: build-image-mockdevices
build-image-mockdevices:
	docker build -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/thermometer/Dockerfile.mockdevice-thermometer --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-thermometer:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/robot-arm/Dockerfile.mockdevice-robot-arm --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-robot-arm:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/plate-reader/Dockerfile.mockdevice-plate-reader --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-plate-reader:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/agv/Dockerfile.mockdevice-agv --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-agv:${IMAGE_VERSION}

.PHONY: buildx-push-image-mockdevices
buildx-push-image-mockdevices:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/thermometer/Dockerfile.mockdevice-thermometer --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-thermometer-multi:${IMAGE_VERSION} --push
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/robot-arm/Dockerfile.mockdevice-robot-arm --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-robot-arm-multi:${IMAGE_VERSION} --push
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/plate-reader/Dockerfile.mockdevice-plate-reader --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-plate-reader-multi:${IMAGE_VERSION} --push
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/examples/mockdevice/agv/Dockerfile.mockdevice-agv --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice-agv-multi:${IMAGE_VERSION} --push


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
	docker rmi $(sudo docker images | grep 'edgehub')
