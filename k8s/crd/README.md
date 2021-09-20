## Commands (Work In Progress)

### Initialize CRD project
```
make kube-builder-init
```

### Create a new kind cluster
```
kind delete cluster && kind create cluster
```

### Install CRD
```
make install
```

### Uninstall CRD
```
make uninstall
```

### Create EdgeDevice API
```
make kube-builder-create-api-edgedevice
```

### Build and publish EdgeDevice controller
```
make docker-build docker-push IMG=edgehub/edgedevice-controller:v0.0.1
```

### Deploy EdgeDeivce controller
```
make deploy IMG=edgehub/edgedevice-controller:v0.0.1
```