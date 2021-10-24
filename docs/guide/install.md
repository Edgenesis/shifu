## via GitHub repository
1. Clone the Shifu repo via:
   ```
   git clone https://github.com/Edgenesis/shifu.git
   ```
2. Install via:
   ```
   
   ```

## via pre-build package
You can install Shifu on an existing Kubernetes cluster via:
```
$ wget --no-check-certificate -O shifu-setup.tar.gz \
        https://edgenesis.com/wp-content/uploads/2021/10/shifu-setup.tar.gz \
        && tar -zxvf shifu-setup.tar.gz && cd shifu
$ bash ./test/scripts/deviceshifu-install.sh apply
```