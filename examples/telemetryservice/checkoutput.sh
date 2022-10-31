#!bin/bash

default='testData'
tdengineOutput=1
# testMQTT
for i in {1..100} 
do
    docker exec nginx curl localhost:9090/mqtt
    output=$(docker exec nginx curl localhost:17773/data)
    echo $output
    echo $default
    if [[ $output == $default ]]
    then
        break
    elif [[ $i == 100 ]]
    then
        exit 1
    fi
done
# init TDEngine Table
docker exec tdengine taos -f /root/init.sql
# testTDEngine
for i in {1..100}
do
    docker exec nginx curl localhost:9090/sql
    output=$(docker exec tdengine taos -s "Select rawdata from shifu.testsubtable limit 1;" | grep 'status api' | wc -l)
    echo $output
    if [[ $output == $tdengineOutput ]]
    then
        exit 0
    fi
done
exit 1
