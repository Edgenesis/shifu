#!/bin/bash
MySQLOutput=2
echo "Waiting for MySQL to be ready..."
while [[ "$(docker inspect --format='{{.State.Health.Status}}' mysql)" != "healthy" ]]; do
  sleep 5
done
echo "MySQL is ready."

# init MySQL Table
docker exec mysql mysql -u root \
    -e "Create database shifu;
        Use shifu;
        CREATE TABLE testTable ( TelemetryID INT AUTO_INCREMENT PRIMARY KEY, DeviceName VARCHAR(255), TelemetryData TEXT, TelemetryTimeStamp DATETIME );
        Select * From testTable;" 

for i in {1..30}
do
    docker exec nginx curl 127.0.0.1:9090/mysql
    output=$(docker exec mysql mysql -u root -e "Use shifu;Select TelemetryData from testTable where TelemetryData='testData' limit 10;" | grep 'testData' | wc -l)
    echo $output
    if [[ $output -ge $MySQLOutput ]]
    then
        exit 0
    fi
done
exit 1
