# 快速上手：安装

## 依赖项
Shifu需要以下依赖项：
1. [Golang](https://golang.org/dl/): Golang是Shifu的开发语言。
2. [Docker](https://docs.docker.com/get-docker/): Shifu的各项服务以Docker镜像的形式存在。
3. [kind](https://kubernetes.io/docs/tasks/tools/): Kind用于以Docker的方式运行本地的Kubernetes集群。
4. [kubectl](https://kubernetes.io/docs/tasks/tools/): Kubernetes的操作工具。
5. [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder): Kubebuilder用于安装CRD。

## 快速配置
Shifu提供了`shifu_install.yml`文件，可以用于快速安装：

```
cd shifu
kubectl apply -f k8s/crd/install/shifu_install.yml
```

## 分步操作
也可以按照如下步骤进行安装：
```
1. 初始化CRD
cd shifu/k8s/crd
make kube-builder-init

2. 创建新集群
// kind delete cluster (in case you have any active kind clusters)
kind create cluster

3. 安装CRD
make install
```

安装Shifu成功后，就可以接入新设备了。



