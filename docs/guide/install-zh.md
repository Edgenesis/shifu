# 安装 Shifu
## 安装/卸载
### 安装
1. 从GitHub 克隆项目到本地：
   ```
   git clone https://github.com/Edgenesis/shifu.git
   ```
2. 执行安装命令：
   ```
   cd shifu
   kubectl apply -f k8s/crd/install/shifu_install.yml
   ```

### 卸载：
1. 执行卸载命令：
   ``` 
   kubectl delete -f k8s/crd/install/shifu_install.yml
   ```
