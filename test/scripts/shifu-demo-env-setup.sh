#!/bin/sh

usage ()
{
  echo "usage: $0 apply/delete"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        if [ "$1" == "apply" ]; then
                Images=("nginx:1.21" "gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0" "edgehub/mockdevice-agv:v0.0.1"
                    "edgehub/mockdevice-plate-reader:v0.0.1" "edgehub/mockdevice-robot-arm:v0.0.1"
                    "edgehub/mockdevice-thermometer:v0.0.1" "edgehub/deviceshifu-http:v0.0.1"
                    "edgehub/edgedevice-controller:v0.0.1")
                kind delete cluster && kind create cluster
                for image in ${Images[*]}; do
                    docker pull $image
                    kind load docker-image $image
                done

                (cd k8s/crd && make install)
                (cd k8s/crd && make deploy IMG=edgehub/edgedevice-controller:v0.0.1)

                kubectl create ns devices
                kubectl config set-context --current --namespace=default
                kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-serviceaccount.yaml
                kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-clusterrole.yaml
                kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-crb.yaml
        else
                kind delete cluster
                docker rmi $(docker images | grep 'edgehub/mockdevice' | awk '{print $3}')
                docker rmi $(docker images | grep 'edgehub/deviceshifu-http' | awk '{print $3}')
                docker rmi $(docker images | grep 'edgehub/edgedevice-controller' | awk '{print $3}')
                docker rmi gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
                docker rmi $(docker images | grep 'kindest/node' | awk '{print $3}')
                docker rmi nginx:1.21
        fi
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi
