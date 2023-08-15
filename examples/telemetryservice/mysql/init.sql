Create database shifu;
Use shifu;
CREATE TABLE testTable ( TelemetryID INT IDENTITY(1,1) PRIMARY KEY, DeviceName VARCHAR(255), TelemetryData TEXT, TelemetryTimeStamp DATETIME );
Select * From testTable;