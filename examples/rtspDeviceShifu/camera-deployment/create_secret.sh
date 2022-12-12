kubectl delete secret -n deviceshifu deviceshifu-secret
kubectl create secret generic deviceshifu-secret --from-literal=rtsp_password=new_password -n deviceshifu
