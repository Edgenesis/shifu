kubectl delete secret -n deviceshifu deviceshifu-secret
kubectl create secret generic deviceshifu-secret --from-literal=opcua_password=new_pwd1 --from-literal=telemetry_service_http_pwd=httppwd --from-literal=telemetry_service_sql_pwd=sqlpwd -n deviceshifu
