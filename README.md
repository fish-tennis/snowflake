# snowflake
[![Go Report Card](https://goreportcard.com/badge/github.com/fish-tennis/snowflake)](https://goreportcard.com/report/github.com/fish-tennis/snowflake)
[![Go Reference](https://pkg.go.dev/badge/github.com/fish-tennis/snowflake.svg)](https://pkg.go.dev/github.com/fish-tennis/snowflake)
[![codecov](https://codecov.io/gh/fish-tennis/snowflake/branch/master/graph/badge.svg?token=L7U3HSD1FV)](https://codecov.io/gh/fish-tennis/snowflake)

an implement of snowflake by go, use atomic instead of mutex

## usage
```go
workerId := uint16(123)
snowFlake := NewSnowFlake(workerId)
theUniqueId := snowFlake.NextId()
```