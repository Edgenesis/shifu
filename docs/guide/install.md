# Install Shifu
## Install/Uninstall
### Install
1. Clone Shifu repository:
   ```
   git clone https://github.com/Edgenesis/shifu.git
   ```
2. Install via:
   ```
   cd shifu
   kubectl apply -f pkg/k8s/crd/install/shifu_install.yml
   ```

### Uninstall
1. To uninstall, issue:
   ```
   kubectl delete -f pkg/k8s/crd/install/shifu_install.yml
   ```

### About Telemetry
To learn more including how to disable the built-in telemetry, checkout out our guide [here!](telemetry.md)