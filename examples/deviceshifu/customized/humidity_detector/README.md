## How to run this example

#### build humidity detector
```bash
docker buildx build -t humidity-detector:nightly .  
kind load docker-image humidity-detector:nightly

docker buildx build -f sample_deviceshifu_dockerfiles/Dockerfile.deviceshifuHTTP-Python \
    -t deviceshifu-http-http-python:nightly ../../../../
kind load docker-image deviceshifu-http-http-python:nightly
```

### Deploy deviceshifu and humidity-detector
```bash
kubectl delete -f configuration
kubectl apply -f configuration
```