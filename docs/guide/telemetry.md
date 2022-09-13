# Telemetry

Telemetry is enabled by default when you install Shifu, while you also have the option to disable it either before or after the installation.

## Data we collect

- External network IP
- Download date
- Kubernetes version
- Shifu version
- Kubernetes cluster size
- Kubernetes Pod Name 
- Kubernetes Deployment Name
- The type of the operating system

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
