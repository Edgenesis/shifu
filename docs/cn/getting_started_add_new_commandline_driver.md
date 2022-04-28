# Getting started: add new commandline driver
If a device communicates via direct command calling executables in the driver, Shifu allows user to add it to the platform, and then this device becomes a Shifu-native device.

## HTTP to SSH driver stub
Shifu has the HTTP to SSH driver stub component. Every time an HTTP request comes, Shifu will ask the stub to convert the request to the actual command and remotely ssh to the place where the driver locates and execute the command.

## Configuration
In a deployment configuration, Shifu needs user to provide the following information:
- **EDGEDEVICE_DRIVER_SSH_KEY_PATH**：The key path of SSH key on driver container which we used to connect to the driver container itself.
- **EDGEDEVICE_DRIVER_HTTP_PORT**(Optional): The HTTP server port of the driver container, default to `11112`.
- **EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND**(Optional): The timeout of an execution.
- **EDGEDEVICE_DRIVER_SSH_USER**(Optional): This is the user we used to SSH into the driver container, default to `root`.

Here is an example of a deployment config:

```
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    app: driver
  name: driver
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: driver
  template:
    metadata:
      labels:
        app: driver
    spec:
      containers:
      - image: http2ssh-stub:v0.0.1
        name: driver
        ports:
        - containerPort: 11112
        env:
        - name: EDGEDEVICE_DRIVER_SSH_KEY_PATH
          value: "/etc/ssh/ssh_host_rsa_key"
        - name: EDGEDEVICE_DRIVER_HTTP_PORT
          value: "11112"
        - name: EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND
          value: "5"
        - name: EDGEDEVICE_DRIVER_SSH_USER
          value: "root"
      - image: nginx:1.21
        name: nginx
```
