#!bin/bash

default='FreeOpcUa Python Server'

for i in {1..5} 
do
    out=(shell kubectl exec -it -n deviceshifu nginx -- curl deviceshifu-opcua/get_server  --connect-timeout 5)

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