### 1. 运行 *Shifu* 并连接一个虚拟traffic light
在 `shifu` 根目录下，运行下面命令来运行 *shifu* :

```shell
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```
在 `shifu` 根目录下，运行下面命令来把虚拟 traffic light 打包成 Docker 镜像:

```shell
docker build -t trafficlight-device:v0.0.1 . -f pkg/fsm/examples/traffic-light-device/Dockerfile
```
在 `shifu` 根目录下，运行下面命令来把虚拟 traffic light 镜像加载到 Kind 中，并部署到 Kubernetes 集群中:

```shell
kind load docker-image trafficlight-device:v0.0.1
kubectl apply -f pkg/fsm/examples/traffic-light-device/configuration
```

### 2. 与 *deviceShifu* 交互
我们可以通过 nginx 应用来和 *deviceShifu* 交互，命令为：

```shell
kubectl run nginx --image=nginx
kubectl exec -it nginx -- bash
```
在 nginx 命令行中通过如下命令与虚拟 traffic light 进行交互：

```shell
curl "trafficlight.devices.svc.cluster.local:11111/get_color";echo 
curl "trafficlight.devices.svc.cluster.local:11111/stop";echo 
curl "trafficlight.devices.svc.cluster.local:11111/proceed";echo 
curl "trafficlight.devices.svc.cluster.local:11111/caution";echo
```