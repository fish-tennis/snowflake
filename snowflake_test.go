package snowflake

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	workerId := uint16(123)
	snowFlake := NewSnowFlake(workerId)
	testId := snowFlake.NextId()
	t.Log(GetTimestampFromId(testId))
	t.Log(GetWorkerIdFromId(testId))
	t.Log(GetSequenceFromId(testId))
	var checkMap sync.Map
	idCount := int64(0)
	for i := 0; i < 16; i++ {
		go func() {
			for {
				atomic.AddInt64(&idCount, 1)
				id := snowFlake.NextId()
				// check duplicate id
				_, loaded := checkMap.LoadOrStore(id, true)
				if loaded {
					t.Errorf("duplicate id:%v", id)
				}
			}
		}()
	}
	tick := time.After(time.Second)
	<-tick
	t.Logf("total:%v", atomic.LoadInt64(&idCount))
}

func BenchmarkAtomic(b *testing.B) {
	runtime.GOMAXPROCS(runtime.NumCPU())
	workerId := uint16(123)
	snowFlake := NewSnowFlake(workerId)
	idCount := int64(0)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			atomic.AddInt64(&idCount, 1)
			snowFlake.NextId()
		}
	})
	tick := time.After(time.Second)
	<-tick
	b.Logf("total:%v", atomic.LoadInt64(&idCount))
}
