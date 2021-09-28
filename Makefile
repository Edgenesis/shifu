PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE_VERSION = v0.0.1

.PHONY: build-image-deviceshifu
build-image-deviceshifu:
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http:${IMAGE_VERSION}

.PHONY: build-image-mockdevices
build-image-mockdevices:
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_thermometer --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_thermometer:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_robot_arm --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_robot_arm:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_tecan --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_tecan:${IMAGE_VERSION}
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.mockdevice_agv --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/mockdevice_agv:${IMAGE_VERSION}

docker-push-image-deviceshifu:
	docker push edgehub/deviceshifu-http:${IMAGE_VERSION}

docker-push-image-mockdevices:
	docker push edgehub/mockdevice_thermometer:${IMAGE_VERSION}
	docker push edgehub/mockdevice_robot_arm:${IMAGE_VERSION}
	docker push edgehub/mockdevice_tecan:${IMAGE_VERSION}
	docker push edgehub/mockdevice_agv:${IMAGE_VERSION}

.PHONY: clean-images
clean-images:
	docker rmi $(sudo docker images | grep 'edgehub')
