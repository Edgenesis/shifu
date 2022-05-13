# 安装 Shifu
## 安装/卸载
### 安装
<<<<<<< HEAD
1. 从 GitHub 克隆项目到本地：
=======
1. 从GitHub 克隆项目到本地：
>>>>>>> 22aa156 (Modify some old documents)
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
