#!bin/bash

default='{"mqtt_message":"","mqtt_receive_timestamp":"0001-01-01 00:00:00 +0000 UTC"}'


out=shell kubectl exec -it -n deviceshifu nginx -- curl deviceshifu-mqtt/mqtt_data

if [[ $out == $default ]]; then
        echo "equal"
    exit 1
fi
echo "not equal"