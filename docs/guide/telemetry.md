# Telemetry

Telemetry is enabled by default when you install Shifu, while you also have the option to disable it either before or after the installation.

## Permissions for telemetry
Shifu's telemetry module leverages Kubernetes' built-in `view` ClusterRole, as detailed in the [official Kubernetes documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/#user-facing-roles).

Telemetry only allows read-only access to most objects, such as Pod basic information, Kubernetes information, and so on. It does not allow access to private information such as roles, secrets, etc., so you don't need to worry about privacy leaks.

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

You can modify telemetry interval by edit `--enable-telemetry` on `pkg/k8s/crd/install/shifu_install.yaml` manually.

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
      image: quay.io/brancz/kube-rbac-proxy:v0.12.0
      name: kube-rbac-proxy
      - args:
        - --telemetry-interval=60 ## Edit this line
```
## To turn-off Telemetry

If you want to turn off temeletry, please delete `--enable-telemetry` on `pkg/k8s/crd/install/shifu_install.yaml` manually.

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
      image: quay.io/brancz/kube-rbac-proxy:v0.12.0
      name: kube-rbac-proxy
      - args:
        - --enable-telemetry ## delete on demand
```
