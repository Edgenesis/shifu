#!/bin/sh

usage ()
{
  echo "usage: $0 apply/delete"
  exit
}

if [ "$1" == "apply" ] || [ "$1" == "delete" ]; then
        if [ "$1" == "apply" ]; then
                kubectl config set-context --current --namespace=default
        fi
        kubectl "$1" -f pkg/k8s/crd/install/shifu_install.yml
else
        echo "not a valid argument, need to be apply/delete"
        exit 0
fi
