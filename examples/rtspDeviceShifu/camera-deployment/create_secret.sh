kubectl delete secret -n deviceshifu rtsp-secret
kubectl create secret generic rtsp-secret --from-literal=rtsp_password=new_password -n deviceshifu
