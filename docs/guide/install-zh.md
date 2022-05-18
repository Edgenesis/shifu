# 安装 Shifu
## 安装/卸载
### 从 GitHub
1. 克隆 GitHub 项目到本地：
   ```
   git clone https://github.com/Edgenesis/shifu.git
   ```
2. 安装：
   ```
   cd shifu
   kubectl apply -f k8s/crd/install/shifu_install.yml
   ```

卸载：
```
kubectl delete -f k8s/crd/install/shifu_install.yml
```
