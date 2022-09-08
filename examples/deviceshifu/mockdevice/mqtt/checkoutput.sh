#!bin/bash

default='{"mqtt_message":"","mqtt_receive_timestamp":"0001-01-01 00:00:00 +0000 UTC"}'

for i in {1..5} 
do
    out=(shell kubectl exec -it -n deviceshifu nginx -- curl deviceshifu-mqtt/mqtt_data --connect-timeout 5)

    if [[ $out == "" ]]
    then 
        echo "empty reply"
    elif [[ $out != $default ]]
    then
        echo "not euqal"
        exit 0
    else 
        exit 1
    fi
done
exit 1