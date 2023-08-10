## Shema for SQL
create table before using telemetryService
### MySQL
```sql
create table yourTableName (
    ts timestamp,
    rawData varchar(255)
);
```

### SQLServer
```sql
create table yourTableName (
    ts datetime2,
    rawData varchar(255)
)
```

### TDengine
```sql
create STable yourTableName (ts TIMESTAMP, rawData varchar(255)) TAGS (defaultTag varchar(255));
```