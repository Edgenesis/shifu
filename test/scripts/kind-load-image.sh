#!/bin/bash

IMAGE_TO_BE_LOADED=(quay.io/brancz/kube-rbac-proxy/kube-rbac-proxy:v0.12.0 edgehub/shifu-controller:v0.0.1)

for img in "${IMAGE_TO_BE_LOADED[@]}"; do
    docker pull $img
    kind load docker-image $img
done
