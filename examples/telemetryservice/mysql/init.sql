Create database shifu;
Use shifu;
CREATE TABLE testTable ( TelemetryID INT AUTO_INCREMENT PRIMARY KEY, DeviceName VARCHAR(255), TelemetryData TEXT, TelemetryTimeStamp DATETIME );
Select * From testTable;