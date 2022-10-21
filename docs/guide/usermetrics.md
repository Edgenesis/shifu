# User Metrics

User Metrics Collection is enabled by default when you install Shifu, while you also have the option to disable it either before or after the installation.

## Permissions for user metrics
Shifu's user metrics collection module leverages Kubernetes' built-in `view` ClusterRole, as detailed in the [official Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles).

User Metrics Collection only allows read-only access to most objects, such as Pod basic information, Kubernetes information, and so on. It does not allow access to private information such as roles, secrets, etc., so you don't need to worry about privacy leaks.

## Data we collect

- External network IP
- Download date
- Kubernetes version
- Shifu version
- Kubernetes cluster size
- Kubernetes Pod Name 
- Kubernetes Deployment Name
- The type of the operating system

## Setting

You can modify user metrics collection interval by edit `--user-metrics-interval=60` on `pkg/k8s/crd/install/shifu_install.yaml` manually.

Or you can also edit via `kubectl edit deployment -n shifu-crd-system shifu-crd-controller-manager` after installation
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
      image: quay.io/brancz/kube-rbac-proxy:v0.13.1
      name: kube-rbac-proxy
      - args:
        - --user-metrics-interval=60 ## Edit this line
```
## To turn-off user-metrics-collection

If you want to turn off user metrics collection, please delete `--enable-user-metrics` on `pkg/k8s/crd/install/shifu_install.yaml` manually.

Or you can also edit via `kubectl edit deployment -n shifu-crd-system shifu-crd-controller-manager` after installation

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
      image: quay.io/brancz/kube-rbac-proxy:v0.13.1
      name: kube-rbac-proxy
      - args:
        - --enable-user-metrics ## delete on demand
```
