package snowflake

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	workerId := uint16(123)
	snowFlake := NewSnowFlake(workerId)
	var checkMap sync.Map
	idCount := int64(0)
	for i := 0; i < 16; i++ {
		go func() {
			for {
				atomic.AddInt64(&idCount,1)
				id := snowFlake.NextId()
				// 检查id是否有重复的值
				_,loaded := checkMap.LoadOrStore(id,true)
				if loaded {
					t.Errorf("duplicate id:%v", id)
				}
			}
			//println(id)
		}()
	}
	tick := time.After(time.Second)
	<-tick
	t.Logf("total:%v", idCount)
}