#!/bin/bash

usage ()
{
  echo "usage: $0 apply/delete"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-sample-edgedevice-serviceaccount.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-sample-edgedevice-clusterrole.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-sample-edgedevice-crb.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-mockdevice-configmap.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-sample-edgedevice-deployment.yaml
        kubectl "$1" -f deviceshifu/examples/mockdevice/test-sample-edgedevice-service.yaml
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi
