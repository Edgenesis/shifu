PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE_VERSION = v0.0.1

.PHONY: build-image-deviceshifu
build-image-deviceshifu:
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http:${IMAGE_VERSION}

.PHONY: buildx-push-image-deviceshifu
buildx-push-image-deviceshifu:
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-multi:${IMAGE_VERSION} --push
	docker buildx build --platform=linux/amd64,linux/arm64,linux/arm -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifuSocket --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-socket:${IMAGE_VERSION} --push

buildx-load-image-deviceshifu:
	docker buildx build --platform=linux/amd64 -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-http:${IMAGE_VERSION} --load
	docker buildx build --platform=linux/amd64 -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifuSocket --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http-socket:${IMAGE_VERSION} --load

.PHONY: download-demo-files
download-demo-files:
	mkdir -p build_dir
	docker pull edgehub/mockdevice-agv:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-plate-reader:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-robot-arm:${IMAGE_VERSION}
	docker pull edgehub/mockdevice-thermometer:${IMAGE_VERSION}
	docker pull edgehub/deviceshifu-http:${IMAGE_VERSION}
	docker pull edgehub/edgedevice-controller:${IMAGE_VERSION}
	docker pull quay.io/brancz/kube-rbac-proxy:v0.8.0
	docker pull kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
	docker pull nginx:1.21

compress-demo-files:
	mkdir -p build_dir
	docker save quay.io/brancz/kube-rbac-proxy:v0.8.0 | gzip > build_dir/kube-rbac-proxy.tar.gz
	docker save edgehub/mockdevice-agv:${IMAGE_VERSION} | gzip > build_dir/mockdevice-agv.tar.gz
	docker save edgehub/mockdevice-plate-reader:${IMAGE_VERSION} | gzip > build_dir/mockdevice-plate-reader.tar.gz
	docker save edgehub/mockdevice-robot-arm:${IMAGE_VERSION} | gzip > build_dir/mockdevice-robot-arm.tar.gz
	docker save edgehub/mockdevice-thermometer:${IMAGE_VERSION} | gzip > build_dir/mockdevice-thermometer.tar.gz
	docker save edgehub/deviceshifu-http:${IMAGE_VERSION} | gzip > build_dir/deviceshifu-http.tar.gz
	docker save edgehub/edgedevice-controller:${IMAGE_VERSION} | gzip > build_dir/edgedevice-controller.tar.gz
	docker save kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6 | gzip > build_dir/kind-image.tar.gz
	docker save nginx:1.21 | gzip > build_dir/nginx.tar.gz
	(cd k8s/crd && make generate-controller-yaml IMG=edgehub/edgedevice-controller:v0.0.1)

compress-edgenesis-files:
	mkdir -p build_dir
	docker save edgehub/mockdevice-agv:${IMAGE_VERSION} | gzip > build_dir/mockdevice-agv.tar.gz
	docker save edgehub/mockdevice-plate-reader:${IMAGE_VERSION} | gzip > build_dir/mockdevice-plate-reader.tar.gz
	docker save edgehub/mockdevice-robot-arm:${IMAGE_VERSION} | gzip > build_dir/mockdevice-robot-arm.tar.gz
	docker save edgehub/mockdevice-thermometer:${IMAGE_VERSION} | gzip > build_dir/mockdevice-thermometer.tar.gz
	docker save edgehub/deviceshifu-http:${IMAGE_VERSION} | gzip > build_dir/deviceshifu-http.tar.gz
	docker save edgehub/edgedevice-controller:${IMAGE_VERSION} | gzip > build_dir/edgedevice-controller.tar.gz

.PHONY: build-deviceshifu-demo-image
build-deviceshifu-demo-image:
	docker build -f ${PROJECT_ROOT}/Dockerfile.demo --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/demo-image-alpine:${IMAGE_VERSION}

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
	docker push edgehub/deviceshifu-http:${IMAGE_VERSION}

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
