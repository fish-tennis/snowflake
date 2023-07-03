# snowflake
an implement of snowflake by go, use atomic instead of mutex

## usage
```go
workerId := uint16(123)
snowFlake := NewSnowFlake(workerId)
theUniqueId := snowFlake.NextId()
```