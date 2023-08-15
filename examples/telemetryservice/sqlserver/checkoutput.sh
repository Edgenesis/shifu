#!bin/bash
SQLServerOutput=2
SQLServerPassword=SomethingComplicated
for i in {1..30}
do
    output=$(docker exec -it sqlserver /opt/mssql-tools/bin/sqlcmd  \
    -S localhost -U sa -P $SQLServerPassword \
    -Q "select name from sys.databases" | grep 'failed' | wc -l)
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
docker exec -it sqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U SA -P $SQLServerPassword -i /root/init.sql

for i in {1..30}
do
    docker exec nginx curl localhost:9090/sql/sqlserver
    output=$(docker exec -it sqlserver /opt/mssql-tools/bin/sqlcmd  \
    -S localhost -U sa -P $SQLServerPassword \
    -Q "use shifu" -Q "SELECT TOP 10 TelemetryData FROM testTable WHERE CAST(TelemetryData AS VARCHAR(MAX)) = 'testData'" | grep 'testData' | wc -l)
    echo $output
    if [[ $output -ge $SQLServerOutput ]]
    then
        exit 0
    fi
done
exit 1
