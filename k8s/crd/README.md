## Commands (Work In Progress)

### Initialize CRD project
```
make kube-builder-init
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