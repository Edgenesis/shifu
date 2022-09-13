# 遥测

安装 Shifu 时默认启用遥测，您可以在安装之前或之后禁用它。

## 我们收集的数据

- 外网IP
- 下载日期
- Kubernetes 版本
- Shifu 版本
- Kubernetes 集群规模
- Kubernetes Pod 名称
- Kubernetes Deployment 名称
- 操作系统的类型

## 关闭遥测

如果要关闭 telemetry，请手动删除 `pkg/k8s/crd/install/shifu_install.yaml` 上的 `--enable-telemetry`。
或者您也可以在安装后通过 `kubectl edit deployment -n shifu-crd-system shifu-crd-controller-manager` 进行编辑

```yaml
apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    spec:
      containers:
      image: quay.io/brancz/kube-rbac-proxy:v0.12.0
      name: kube-rbac-proxy
      - args:
        - --enable-telemetry ## delete on demand
```