#!bin/bash
MySQLOutput=2
for i in {1..30}
do
    output=$(docker exec mysql mysql -h localhost -u root -p"" -e "Show databases;" | grep 'error' | wc -l)
    echo $output
    if [[ $output -eq 0 ]]
    then
        break
    elif [[ $i -eq 30 ]]
    then
        exit 1
    fi
done
# init MySQL Table
docker exec mysql mysql -h localhost -u root -p"" < /root/init.sql

for i in {1..30}
do
    docker exec nginx curl localhost:9090/sql/mysql
    output=$(docker exec mysql mysql -h localhost -u root -p"" -e "Use shifu;" -e "Select TelemetryData from testTable where TelemetryData='testData' limit 10;" | grep 'testData' | wc -l)
    echo $output
    if [[ $output -ge $MySQLOutput ]]
    then
        exit 0
    fi
done
exit 1
