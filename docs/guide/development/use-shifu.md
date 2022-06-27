# User Guide

Hello developers! This guide will help you control the virtual device `demo` locally using `Shifu`.

### This guide has been tested on the following platforms:
```
Windows 10
```

If you have any problems using this guide and find any bugs, please don't hesitate to open an [issue] in GitHub. The address is : https://github.com/Edgenesis/shifu/issues

# step:
## 1. Use `kind` to deploy the cluster locally
Create a `k8s` cluster locally with the following command:
```sh
kind create cluster
```

## 2. View all pods in the cluster
Run the following command to view all pods in the current cluster:
```sh
kubectl get pods -A
```

## 3. Deploy `shifu` to the `k8s` cluster
Run the following command to deploy shifu to the cluster:
```shell
kubectl apply -f k8s/crd/install/shifu_install.yml
```

## 4. Confirm that `shifu` has been added to the `k8s` cluster
Run the following command to view all pods in the current cluster and confirm that shifu has been deployed to the cluster:
```shell
kubectl get pods -A
```

## 5. Generate a digital twin of the `demo` device
Run the following command to create a digital twin of the `demo` device:
```shell
kubectl apply -f deviceshifu/examples/demo_device/edgedevice-agv/
```

## 6. Deploy and run the `nginx` application
Run the `nginx` application with the following command:
```shell
kubectl run nginx --image=nginx
```

## 7. Enter a command to return to the running state of the `demo` device
Run the following command to get the real-time coordinates of the `demo` device:
```shell
kubectl exec -it nginx -- bash
curl deviceshifu-agv.deviceshifu.svc.cluster.local/get_position;echo
```


