# snowflake
an implement of snowflake by go, use atomic instead of mutex

雪花算法的一种go实现,使用原子锁(atomic),取代互斥锁(mutex)

## usage
```go
workerId := uint16(123)
snowFlake := NewSnowFlake(workerId)
theUniqueId := snowFlake.NextId()
```