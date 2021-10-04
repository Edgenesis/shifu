PROJECT_ROOT := $(abspath $(dir $(lastword $(MAKEFILE_LIST))))
IMAGE_VERSION = v0.0.1

.PHONY: build-image-deviceshifu
build-image-deviceshifu:
	docker build -f ${PROJECT_ROOT}/deviceshifu/Dockerfile.deviceshifu --build-arg PROJECT_ROOT="${PROJECT_ROOT}" ${PROJECT_ROOT} -t edgehub/deviceshifu-http:${IMAGE_VERSION}

docker-push-image-deviceshifu:
	docker push edgehub/deviceshifu-http:${IMAGE_VERSION}

.PHONY: clean-images
clean-images:
	docker rmi $(sudo docker images | grep 'edgehub')
