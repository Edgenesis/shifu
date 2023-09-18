#!bin/bash
SQLServerOutput=2
sleep 6
MAX_RETRIES=5
for retry in $(seq 1 $MAX_RETRIES); do
    echo "Attempt $retry of $MAX_RETRIES"

    for i in {1..50}; do
        output=$(docker exec sqlserver /opt/mssql-tools/bin/sqlcmd \
        -S localhost -U sa -P Some_Strong_Password \
        -Q "select name from sys.databases" | grep 'Error' | wc -l)
        echo "Output: $output"
        
        if [[ $output -eq 0 ]]; then
            echo "Database connection successful"
            break
        elif [[ $i -eq 50 ]]; then
            echo "connection failed, try again"
            sleep 5
        fi
    done

    if [[ $output -eq 0 ]]; then
        break
    fi
done

# init SQLServer Table
docker exec sqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U SA -P Some_Strong_Password \
    -Q "Create database shifu;"

docker exec sqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U SA -P Some_Strong_Password \
    -d shifu -Q "CREATE TABLE testTable ( TelemetryID INT IDENTITY(1,1) PRIMARY KEY, DeviceName VARCHAR(255), TelemetryData TEXT, TelemetryTimeStamp DATETIME );"

docker exec sqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U SA -P Some_Strong_Password \
    -d shifu -Q "Select * from shifu.dbo.testTable;"

for i in {1..30}
do
    docker exec nginx curl localhost:9090/sqlserver
    output=$(docker exec sqlserver /opt/mssql-tools/bin/sqlcmd  \
    -S localhost -U sa -P Some_Strong_Password \
    -Q "SELECT TOP 10 TelemetryData FROM shifu.dbo.testTable WHERE CAST(TelemetryData AS VARCHAR(MAX)) = 'testData'" | grep 'testData' | wc -l)
    echo $output
    if [[ $output -ge $SQLServerOutput ]]
    then
        exit 0
    fi
done
exit 1
