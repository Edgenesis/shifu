#!/bin/sh

usage ()
{
  echo "usage: $0 apply/delete"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        if [ "$1" == "apply" ]; then
                (cd k8s/crd && make generate-controller-yaml IMG=edgehub/edgedevice-controller:v0.0.1)
                kubectl create ns devices
                kubectl config set-context --current --namespace=default
        fi
        kubectl "$1" -f k8s/crd
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-serviceaccount.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-clusterrole.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-edgedevice-mockdevice-crb.yaml
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi
