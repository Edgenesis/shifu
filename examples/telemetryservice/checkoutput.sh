#!bin/bash

default='{"mqtt_message":"","mqtt_receive_timestamp":"0001-01-01 00:00:00 +0000 UTC"}'

for i in {1..5} 
do
    kubectl run client --image=edgehub/mockclient:$(tag)

    kubectl 
done
exit 1
