#!/bin/sh

usage ()
{
  echo "usage: $0 apply/delete"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        if [ "$1" == "apply" ]; then
                (cd /build_dir && for f in *.tar.gz; do docker load < $f; done)
                kind delete cluster && kind create cluster
                kind load docker-image nginx:1.21
                kind load docker-image quay.io/brancz/kube-rbac-proxy:v0.12.0
                kind load docker-image edgehub/mockdevice-agv:v0.0.1
                kind load docker-image edgehub/mockdevice-plate-reader:v0.0.1
                kind load docker-image edgehub/mockdevice-robot-arm:v0.0.1
                kind load docker-image edgehub/mockdevice-thermometer:v0.0.1
                kind load docker-image edgehub/deviceshifu-http-http:v0.0.1
                kind load docker-image edgehub/edgedevice-controller-multi:v0.0.1
                kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
        else
                kind delete cluster
                docker rmi $(docker images | grep 'edgehub/mockdevice' | awk '{print $3}')
                docker rmi $(docker images | grep 'edgehub/deviceshifu-http-http' | awk '{print $3}')
                docker rmi $(docker images | grep 'edgehub/edgedevice-controller' | awk '{print $3}')
                docker rmi quay.io/brancz/kube-rbac-proxy:v0.12.0
                docker rmi $(docker images | grep 'kindest/node' | awk '{print $3}')
                docker rmi nginx:1.21
        fi
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi
