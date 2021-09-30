#!/bin/sh

usage ()
{
  echo "usage: $0 apply/delete"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        if [ "$1" == "apply" ]; then
                (cd /build_dir && for f in *.tar; do cat $f | docker load; done)
                kind delete cluster && kind create cluster
                kind load docker-image nginx:1.21
                kind load docker-image gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0
                kind load docker-image edgehub/mockdevice_agv:v0.0.1
                kind load docker-image edgehub/mockdevice_tecan:v0.0.1
                kind load docker-image edgehub/mockdevice_robot_arm:v0.0.1
                kind load docker-image edgehub/mockdevice_thermometer:v0.0.1
                kind load docker-image edgehub/deviceshifu-http:v0.0.1
                kind load docker-image edgehub/edgedevice-controller:v0.0.1
                kubectl apply -f k8s/crd
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
