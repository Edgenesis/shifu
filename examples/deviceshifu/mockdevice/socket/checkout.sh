#!bin/bash

default='{"message":"123\n","status":200}'

for i in {1..5} 
do
    out=(kubectl exec -it -n deviceshifu nginx -- curl -XPOST -H "Content-Type:application/json" deviceshifu-socket.deviceshifu.svc.cluster.local/cmd -d '{"command":"123"}'  --connect-timeout 5)

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