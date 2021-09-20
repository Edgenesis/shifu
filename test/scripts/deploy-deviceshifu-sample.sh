#!/bin/bash

kubectl apply -f deviceshifu/examples/mockdevice/test-sample-edgedevice-serviceaccount.yaml
kubectl apply -f deviceshifu/examples/mockdevice/test-sample-edgedevice-clusterrole.yaml
kubectl apply -f deviceshifu/examples/mockdevice/test-sample-edgedevice-crb.yaml
kubectl apply -f deviceshifu/examples/mockdevice/test-sample-edgedevice-deployment.yaml
kubectl apply -f deviceshifu/examples/mockdevice/test-sample-edgedevice-service.yaml
