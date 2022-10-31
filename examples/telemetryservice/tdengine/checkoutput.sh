#!bin/bash

# init TDEngine Table
docker exec tdengine taos -f /root/init.sql

for i in {1..30}
do
    docker exec nginx curl localhost:9090/sql
    output=$(docker exec tdengine taos -s "Select rawdata from shifu.testsubtable where rawdata='testData' limit 1;" | grep 'testData' | wc -l)
    echo $output
    if [[ $output -ge $tdengineOutput ]]
    then
        exit 0
    fi
done
exit 1
