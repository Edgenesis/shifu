# 使用指南

开发者你好！ 本份指南会帮助你在本地使用 `Shifu` 控制虚拟设备 `demo` 。

### 本指南已在以下平台中测试过:
```
Ubuntu WSL on Windows 10
```

如您在使用本指南中有任何的问题以及发现了任何错误请毫不犹豫的在GitHub中建立一个 [issue](https://github.com/Edgenesis/shifu/issues) 。

# 步骤:
1. 利用`kind`在本地部署集群。
使用以下命令在本地创建一个`k8s`集群:
```sh
kind create cluster
```

2. 查看集群中的所有`pod`。
运行以下命令来查看当前集群中的所有`pod`:
```sh
kubectl get pods -A
```

3. 将`shifu`部署到`k8s`集群中。
在`shifu`根目录中运行以下命令将`shifu`部署到集群中：
```shell
kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
```

4. 确认`shifu`已经被加入到`k8s`集群中。
运行以下命令查看当前集群中的所有`pod`，并确认`shifu`已被部署到集群中：
```shell
kubectl get pods -A
```

5. 生成`demo`设备的数字孪生。
在`shifu`根目录中运行以下命令创建`demo`设备的数字孪生：
```shell
kubectl apply -f examples/deviceshifu/demo_device/edgedevice-agv/
```

6. 部署与运行`nginx`应用。
用以下命令来运行`nginx`应用：
```shell
kubectl run nginx --image=nginx
```

7. 输入一条指令，返回`demo`设备的运行状态。
运行以下命令得到`demo`设备的实时坐标值:
```shell
kubectl exec -it nginx -- bash
curl deviceshifu-agv.deviceshifu.svc.cluster.local/get_position;echo
```