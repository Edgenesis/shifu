## 在本地测试 OPC UA ***deviceShifu***:
1. 在本地的 Kubernetes 集群 (kind/k3d均可) 安装 Shifu
2. 安装 Python 的依赖库:
    - `pip3 install server/requirements.txt`
3. 启动测试的 OPC UA 服务器:
    ```bash
    localhost server % python3 server.py 
    Endpoints other than open requested but private key and certificate are not set.
    Listening on 0.0.0.0:4840
    ```
4. 从 `shifu/examples/opcuaDeviceShifu`中部署 OPC UA ***deviceShifu***:
    ```bash
    # kubectl apply -f opcua_deploy/
    configmap/opcua-configmap-0.0.1 created
    deployment.apps/edgedevice-opcua-deployment created
    service/edgedevice-opcua created
    edgedevice.shifu.edgenesis.io/edgedevice-opcua created
    ```
5. 启动一个 `Nginx` pod 并进入其命令行:
    ```bash
    # kubectl run nginx --image=nginx
    ```
    ```bash
    # kubectl exec -it nginx -- bash
    root@nginx:/#
    ```
6. 和 `OPC UA` ***deviceShifu*** 进行互动:
    ```bash
    root@nginx:/# curl edgedevice-opcua/get_time;echo
    2022-05-25 07:29:36.879869 +0000 UTC
    root@nginx:/# curl edgedevice-opcua/get_server;echo
    FreeOpcUa Python Server
    root@nginx:/# curl edgedevice-opcua/get_value;echo
    3175.5999999982073
    ```
