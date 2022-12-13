kubectl delete secret -n deviceshifu deviceshifu-secret
kubectl create secret generic deviceshifu-secret --from-literal=telemetry_service_sql_pwd=sqlpwd -n deviceshifu
