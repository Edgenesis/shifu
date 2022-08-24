# HTTP to SSH driver stub
## 介绍
为了使 ***Shifu*** 可以整合开发者的驱动，我们编写了一个简单的 HTTP 到 SSH 的 stub来供开发者使用

### 设计
这个 HTTP 到 SSH 的 stub 设计如下：
- 从 stub 利用容器提供的公钥建立一个到容器的 SSH 连接
- 这个 SSH 连接会被用作反向 HTTP 代理，代理到 localhost 的制定 HTTP 端口
- 这个 stub 会直接在 SSH 的会话中执行 HTTP 请求的 body 的内容
- stub 会将执行内容的结果以及状态代理回复给请求者

### 功能
#### 代理 HTTP body 的内容到 SSH shell 并执行
stub 的主要功能就是将任意HTTP 请求中 body 的内容附加一个超时并执行

举例:

当使用 CURL 去 Post 一个请求到制定 URL， 会是如下命令：

`curl -X POST -d "ping 8.8.8.8" http://example.com`

这个请求会被 stub 传递到驱动容器中的 `shell` 中执行：

`~ # ping 8.8.8.8`

HTTP 的返回结果会是如下样式 （注意这个输出不全，是因为我们设定了一个命令超时的环境变量）：

```
PING 8.8.8.8 (8.8.8.8): 56 data bytes
64 bytes from 8.8.8.8: seq=0 ttl=36 time=47.227 ms
64 bytes from 8.8.8.8: seq=1 ttl=36 time=50.137 ms
64 bytes from 8.8.8.8: seq=3 ttl=36 time=47.619 ms
```

#### 检查 `session.Run(cmd)` 错误并设定 HTTP 的返回状态码
当前 stub 会在成功时返回 `200`， 错误或超时时返回 `400`

对于出错的命令，stub 会将 `stdout` 和 `stderr` 合并到一起通过 HTTP 的 body 返回

### 使用
我们写了一个 Dockerfile 的示例在 [`driver_util/examples/simple-alpine/Dockerfile.sample`](/examples/driver_util/simple-alpine/Dockerfile.sample)，其中演示了如何将这个 stub 加入到一个 alpine 的 Docker 镜像中

打包的 Docker 镜像会使用以下变量，因此我们需要在[yaml文件](/examples/driver_utilutil/simple-alpine/driver.yaml)中进行配置：
- `EDGEDEVICE_DRIVER_SSH_KEY_PATH`(必填)
  - 用来建立 SSH 连接到自身的 SSH 密钥路径
- `EDGEDEVICE_DRIVER_HTTP_PORT` (选填)
  - HTTP 服务的端口，默认为 `11112`
- `EDGEDEVICE_DRIVER_EXEC_TIMEOUT_SECOND` (选填)
  - 执行命令的超时，通过在命令前面加上 `timeout <seconds>` 来实现
- `EDGEDEVICE_DRIVER_SSH_USER` (选填)
  - 我们 SSH 到自身容器的用户，默认为 `root`
 