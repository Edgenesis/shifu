#!/bin/sh

usage ()
{
  echo "usage: $0 apply/delete deviceDemo/applicationDemo"
  exit
}

if ([ "$1" == "apply" ] || [ "$1" == "delete" ]) && ([ "$2" == "deviceDemo" ] || [ "$2" == "applicationDemo" ]); then
        if [ "$1" == "apply" ]; then
                Images=("nginx:1.21" "gcr.io/kubebuilder/kube-rbac-proxy:v0.8.0" "edgehub/mockdevice-agv:v0.0.1"
                    "edgehub/mockdevice-plate-reader:v0.0.1" "edgehub/mockdevice-robot-arm:v0.0.1"
                    "edgehub/mockdevice-thermometer:v0.0.1" "edgehub/deviceshifu-http:v0.0.1"
                    "edgehub/edgedevice-controller:v0.0.1")
                kind delete cluster && kind create cluster
                for image in ${Images[*]}; do
                    if [ "$(docker images -q $image 2> /dev/null)" == "" ]; then
                        echo "going to pull docker image $image..."
                        docker pull $image
                    fi
                    kind load docker-image $image
                done

                (cd k8s/crd && make install)
                (cd k8s/crd && make deploy IMG=edgehub/edgedevice-controller:v0.0.1)

                kubectl create ns devices
                kubectl config set-context --current --namespace=default
                if [ "$2" == "deviceDemo" ]; then
                    kubectl "$1" -f k8s/crd/config/samples/shifu_v1alpha1_edgedevice.yaml
                    kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-configmap.yaml
                    kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-deployment.yaml
                    kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-service.yaml
                fi
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
                if [ "$2" == "deviceDemo" ]; then
                    kubectl "$1" -f k8s/crd/config/samples/shifu_v1alpha1_edgedevice.yaml
                    kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-configmap.yaml
                    kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-deployment.yaml
                    kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-service.yaml
                fi
        fi
else
        echo "not a valid argument, need to be apply/delete deviceDemo/applicationDemo"
        exit 0
fi
