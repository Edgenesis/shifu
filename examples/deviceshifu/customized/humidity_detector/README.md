## How to run this example

#### build humidity detector
```bash
docker buildx build -t edgehub/humidity-detector:nightly .  
kind load docker-image edgehub/humidity-detector:nightly

docker buildx build -f sample_deviceshifu_dockerfiles/Dockerfile.deviceshifuHTTP-Python \
    -t edgehub/deviceshifu-http-http-python:nightly ../../../../
kind load docker-image edgehub/deviceshifu-http-http-python:nightly
```

### Deploy deviceshifu and humidity-detector
```bash
kubectl delete -f configuration
kubectl apply -f configuration
```