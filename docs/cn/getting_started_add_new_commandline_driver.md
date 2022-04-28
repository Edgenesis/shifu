# 快速上手: 添加新驱动
如果一个驱动使用命令行可执行文件，Shifu支持将它直接接入，从而可以远程操作使用这类驱动的设备。

## HTTP to SSH driver stub
这个组件的功能是将接收到的HTTP请求转化为对驱动可执行文件的命令并远程执行。

## 配置
首先，我们需要将带有可执行文件的驱动打包成一个Docker容器镜像。

在Deployment的配置中，Shifu需要下列四个元素:
- **EDGEDEVICE_DRIVER_SSH_KEY_PATH**：SSH key 的路径，Shifu需要通过它来访问驱动的容器。
- **EDGEDEVICE_DRIVER_HTTP_PORT**(可选): 驱动容器的，默认值为`11112`.
- **EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND**(可选): 一个操作的超时设置。
- **EDGEDEVICE_DRIVER_SSH_USER**(可选): 驱动容器的用户账户，默认为`root`。

下面是一个简单的例子:

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
