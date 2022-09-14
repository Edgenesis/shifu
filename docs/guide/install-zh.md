# 安装 Shifu
## 安装/卸载
### 安装
1. 从 GitHub 克隆项目到本地:
   ```
   git clone https://github.com/Edgenesis/shifu.git
   ```
2. 执行安装命令:
   ```
   cd shifu
   kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
   ```

### 卸载：
1. 执行卸载命令:
   ``` 
   kubectl delete -f pkg/k8s/crd/install/shifu_install.yml
   ```

### 关于遥测
要了解更多信息，包括如何禁用内置遥测，请在[此处](telemetry-zh.md)查看我们的指南