# Install Shifu
## Install/Uninstall
### Install
<<<<<<< HEAD
1. Clone Shifu repository:
=======
1. Clone Shifu from  Github to local:
>>>>>>> 22aa156 (Modify some old documents)
   ```
   git clone https://github.com/Edgenesis/shifu.git
   ```
2. Install via:
   ```
   cd shifu
   kubectl apply -f k8s/crd/install/shifu_install.yml
   ```

### Uninstall
1. To uninstall, issue:
   ```
   kubectl delete -f k8s/crd/install/shifu_install.yml
   ```
