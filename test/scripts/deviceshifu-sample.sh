#!/bin/bash

usage ()
{
  echo "usage: $0 apply/delete"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        if [ "$1" == "apply" ]; then
                (cd k8s/crd && make install)
                (cd k8s/crd && make deploy IMG=edgehub/edgedevice-controller:v0.0.1)
                kubectl config set-context --current --namespace=default
                kubectl "$1" -f k8s/crd/config/samples/shifu_v1alpha1_edgedevice.yaml
        else
                kubectl "$1" -f k8s/crd/config/samples/shifu_v1alpha1_edgedevice.yaml
        fi
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-serviceaccount.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-clusterrole.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-crb.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-configmap.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-deployment.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-service.yaml
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi
