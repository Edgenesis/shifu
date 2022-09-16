# 遥测

安装 Shifu 时默认启用遥测，您可以在安装之前或之后禁用它。

## 遥测的权限

Shifu 的遥测模块利用了Kubernetes内置的 `view` 权限的ClusterRole，详情请见[Kubernetes 官方文档](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/rbac/#user-facing-roles)。

遥测仅允许对大多数对象有只读权限,例如Pod基本信息、Kubernetes信息等。 它不允许查看角色、Secrets等隐私信息，所以您无需担心隐私泄漏问题。

## 我们收集的数据

- 外网IP
- 下载日期
- Kubernetes 版本
- Shifu 版本
- Kubernetes 集群规模
- Kubernetes Pod 名称
- Kubernetes Deployment 名称
- 操作系统的类型

## 设置

您可以通过设置  `pkg/k8s/crd/install/shifu_install.yaml` 上的 `--telemetry-interval=60` 对遥测的间隔时间进行设置。

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
        - --telemetry-interval=60 ## 编辑此行
```
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
        - --enable-telemetry ## 删除此行
```
