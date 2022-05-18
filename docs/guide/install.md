# Install Shifu
## Install/Uninstall
### via GitHub repository
1. Clone the Shifu repo via:
   ```
   git clone https://github.com/Edgenesis/shifu.git
   ```
2. Install via:
   ```
   cd shifu
   kubectl apply -f k8s/crd/install/shifu_install.yml
   ```

To uninstall, issue:
```
kubectl delete -f k8s/crd/install/shifu_install.yml
```
