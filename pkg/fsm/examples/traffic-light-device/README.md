### 1. Running Shifu and connecting a virtual traffic light
In the shifu root directory, run the following command to run Shifu:
```shell
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```

In the shifu root directory, run the following command to package the virtual traffic light into a Docker image:
```shell
docker build -t trafficlight-device:v0.0.1 . -f pkg/fsm/examples/traffic-light-device/Dockerfile
```

In the shifu root directory, run the following commands to load the virtual traffic light image into Kind and deploy it to the Kubernetes cluster:
```shell
kind load docker-image trafficlight-device:v0.0.1
kubectl apply -f pkg/fsm/examples/traffic-light-device/configuration
```

### 2. Interacting with deviceShifu
We can interact with deviceShifu through the nginx application using the following commands:
```shell
kubectl run nginx --image=nginx
kubectl exec -it nginx -- bash
```

In the nginx command line, use the following commands to interact with the virtual traffic light:
```shell
curl "trafficlight.devices.svc.cluster.local:11111/get_color";echo 
curl "trafficlight.devices.svc.cluster.local:11111/stop";echo 
curl "trafficlight.devices.svc.cluster.local:11111/proceed";echo 
curl "trafficlight.devices.svc.cluster.local:11111/caution";echo
```
These commands allow you to communicate with the virtual traffic light deployed in the Kubernetes cluster by making HTTP requests to the specified endpoints.