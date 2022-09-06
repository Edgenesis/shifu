#!bin/bash

default='{"mqtt_message":"","mqtt_receive_timestamp":"0001-01-01 00:00:00 +0000 UTC"}'

mosquitto_pub -h localhost -d -p 18830 -t /test/test -m "test2333"
kubectl exec -it -n deviceshifu nginx -- curl deviceshifu-mqtt/mqtt_data

if [[ $1 == $default ]]; then
    echo "not equal"
    exit 1
fi