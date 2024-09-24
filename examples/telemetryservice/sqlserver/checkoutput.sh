#!bin/bash
SQLServerOutput=2
echo "Waiting for SQL Server to be ready..."
while [[ "$(docker inspect --format='{{.State.Health.Status}}' sqlserver)" != "healthy" ]]; do
  sleep 5
done
echo "SQL Server is ready."

# init SQLServer Table
docker exec sqlserver /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -C -P YourStrong@Passw0rd \
    -Q "Create database shifu;"

docker exec sqlserver /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -C -P YourStrong@Passw0rd \
    -d shifu -Q "CREATE TABLE testTable ( TelemetryID INT IDENTITY(1,1) PRIMARY KEY, DeviceName VARCHAR(255), TelemetryData TEXT, TelemetryTimeStamp DATETIME );"

docker exec sqlserver /opt/mssql-tools18/bin/sqlcmd -S localhost -U sa -C -P YourStrong@Passw0rd \
    -d shifu -Q "Select * from shifu.dbo.testTable;"

for i in {1..30}
do
    docker exec nginx curl localhost:9090/sqlserver
    output=$(docker exec sqlserver /opt/mssql-tools18/bin/sqlcmd  \
    -S localhost -U sa -C -P YourStrong@Passw0rd \
    -Q "SELECT TOP 10 TelemetryData FROM shifu.dbo.testTable WHERE CAST(TelemetryData AS VARCHAR(MAX)) = 'testData'" | grep 'testData' | wc -l)
    echo $output
    if [[ $output -ge $SQLServerOutput ]]
    then
        exit 0
    fi
done
exit 1
