PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE_VERSION = v0.0.1

.PHONY: build-image-deviceshifu
build-image-deviceshifu:
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http:${IMAGE_VERSION}

.PHONY: download-demo-files
download-demo-files:
	mkdir -p build_dir
	docker pull edgehub/mockdevice_agv:${IMAGE_VERSION}
	docker pull edgehub/mockdevice_tecan:${IMAGE_VERSION}
	docker pull edgehub/mockdevice_robot_arm:${IMAGE_VERSION}
	docker pull edgehub/mockdevice_thermometer:${IMAGE_VERSION}
	docker pull edgehub/deviceshifu-http:${IMAGE_VERSION}
	docker pull edgehub/edgedevice-controller:${IMAGE_VERSION}
	docker pull gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
	docker pull kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6
	docker pull nginx:1.21
	docker save gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0 > build_dir/kube-rbac-proxy.tar
	docker save edgehub/mockdevice_agv:${IMAGE_VERSION} > build_dir/mockdevice_agv.tar
	docker save edgehub/mockdevice_tecan:${IMAGE_VERSION} > build_dir/mockdevice_tecan.tar
	docker save edgehub/mockdevice_robot_arm:${IMAGE_VERSION} > build_dir/mockdevice_robot_arm.tar
	docker save edgehub/mockdevice_thermometer:${IMAGE_VERSION} > build_dir/mockdevice_thermometer.tar
	docker save edgehub/deviceshifu-http:${IMAGE_VERSION} > build_dir/deviceshifu-http.tar
	docker save edgehub/edgedevice-controller:${IMAGE_VERSION} > build_dir/edgedevice-controller.tar
	docker save kindest/node:v1.21.1@sha256:69860bda5563ac81e3c0057d654b5253219618a22ec3a346306239bba8cfa1a6 > build_dir/kind-image.tar
	docker save nginx:1.21 > build_dir/nginx.tar

.PHONY: build-deviceshifu-demo-image
build-deviceshifu-demo-image:
	docker build -f ${PROJECT_ROOT}/Dockerfile.demo --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/demo_image-alpine:${IMAGE_VERSION}

.PHONY: build-image-mockdevices
build-image-mockdevices:
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_thermometer --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_thermometer:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_robot_arm --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_robot_arm:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_tecan --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_tecan:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_agv --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_agv:${IMAGE_VERSION}

docker-push-image-deviceshifu:
	docker push edgehub/deviceshifu-http:${IMAGE_VERSION}

docker-push-deviceshifu-demo-image:
	docker push edgehub/demo_image-alpine:${IMAGE_VERSION}

docker-push-image-mockdevices:
	docker push edgehub/mockdevice_thermometer:${IMAGE_VERSION}
	docker push edgehub/mockdevice_robot_arm:${IMAGE_VERSION}
	docker push edgehub/mockdevice_tecan:${IMAGE_VERSION}
	docker push edgehub/mockdevice_agv:${IMAGE_VERSION}

.PHONY: clean-images
clean-images:
	docker rmi $(sudo docker images | grep 'edgehub')
