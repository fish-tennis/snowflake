package snowflake

import (
	"errors"
	"fmt"
	"sync/atomic"
	"time"
	"unsafe"
)

// uint64的第1位保留符号位,还有63位可用
// 41位时间值 + 13位WorkerId + 9位序号
const (
	// 时间值位数
	bitLenTime = 41 // bit length of time
	// 序号位数
	bitLenSequence = 9 // bit length of sequence number
	// WorkerId位数
	bitLenWorkerId = 63 - bitLenTime - bitLenSequence // bit length of machine id
	// 序号最大值(同一时间周期内最多能生成多少个id)
	maxSequence = (1 << bitLenSequence) - 1 // max sequence number
	// WorkerId最大值
	maxWorkerId = (1 << bitLenWorkerId) - 1
	// WorkerId位移
	shiftWorkerId = bitLenSequence
	// 时间值位移
	shiftTime = bitLenSequence + bitLenWorkerId
	// 时间值开始时间
	startTime = 1609430400000 // 2021-01-01 00:00:00
	//// 最多能用多少年
	//maxYears = (1 << bitLenTime) / (1000 * 60 * 60 * 24 * 365)
	//// 理论上每秒最多可生成多少个id
	//maxIdCountPerSecond = (1 << bitLenSequence) * 1000
)

type SnowFlake struct {
	// 区分进程的id,不同的进程不能重复
	// process id
	workerId uint16
	// 当前时间周期(毫秒)
	// current time cycle (time.Millisecond)
	timeAndSequence *timeCycle
}

// 时间和序号封装为一个struct,以保证数据一致性
// use struct to ensure consistency of many field when use atomic.CompareAndSwapPointer
type timeCycle struct {
	sequence int32
	time     int64 // time.Millisecond
}

// workerId:[0,maxWorkerId]
func NewSnowFlake(workerId uint16) *SnowFlake {
	if workerId > maxWorkerId {
		panic(errors.New(fmt.Sprintf("workerId:[0,%v]", maxWorkerId)))
	}
	return &SnowFlake{
		workerId: workerId,
		timeAndSequence: &timeCycle{
			sequence: 0,
			time:     time.Now().UnixNano() / int64(time.Millisecond),
		},
	}
}

// 生成唯一id
// generate unique id
func (sf *SnowFlake) NextId() uint64 {
	for {
		curTimeCycle := (*timeCycle)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&sf.timeAndSequence))))
		curTime := curTimeCycle.time
		newSequence := int32(0)
		now := time.Now().UnixNano() / int64(time.Millisecond)
		if now == curTimeCycle.time {
			// 同一个时间周期内,自增序列号
			// increase sequence in the same timeCycle
			newSequence = atomic.AddInt32(&curTimeCycle.sequence, 1)
			if newSequence > maxSequence {
				// 当前事件周期的序列号已用完,等待下一个事件周期
				time.Sleep(time.Millisecond) // wait for the next timeCycle
				continue
			}
		} else {
			// 时间值必须是越来越大的
			// new timeCycle'time must greater than old
			if now < curTime {
				println(fmt.Sprintf("now(%v) < curTime(%v))", now, curTime))
				// 有可能手动设置了系统时间,导致时间后退了
				// It is possible that the system time was manually set, causing the time to go back
				if curTime-now <= int64(time.Second/time.Millisecond) {
					// 如果时间回退在1秒钟之内,则等待时间追上
					// If the time goes back less than 1 second, sleep for awhile
					time.Sleep(time.Duration(curTime-now))
					continue
				}
				// 时间回退超过1秒钟,则继续运行,防止"死循环",但是后面生成的id不排除有重复值
				// If the time goes back more than 1 second, continue running to prevent a "dead loop",
				// but the id generated later may not unique
			}
			newTimeCycle := &timeCycle{
				sequence: 0,
				time:     now,
			}
			// 当同一个时间周期内有多个协程并发时,只有一个协程能成功更新时间周期
			// only one routine can update the timeCycle value at same time
			if atomic.CompareAndSwapPointer((*unsafe.Pointer)(unsafe.Pointer(&sf.timeAndSequence)), unsafe.Pointer(curTimeCycle), unsafe.Pointer(newTimeCycle)) {
				newSequence = 0
				curTime = newTimeCycle.time
				//println("newTimeCycle", newTimeCycle.time)
			} else {
				// 虽然该协程没成功更新时间周期,但是如果上个时间周期内的序号还没用完,那就可以生成id
				// if the old timeCycle's sequence also enough, use it
				newSequence = atomic.AddInt32(&curTimeCycle.sequence, 1)
				if newSequence > maxSequence {
					// 当前事件周期的序列号已用完,等待下一个事件周期
					time.Sleep(time.Millisecond) // wait for the next timeCycle
					continue
				}
			}
		}
		// time | WorkerId | sequence
		return (uint64(curTime-startTime) << shiftTime) | (uint64(sf.workerId) << shiftWorkerId) | uint64(newSequence)
	}
}

// 获取id的生成时间戳(毫秒)
// time.Millisecond
func GetTimestampFromId(snowflakeId uint64) int64 {
	return int64((snowflakeId >> shiftTime) + startTime)
}

func GetWorkerIdFromId(snowflakeId uint64) uint16 {
	return uint16((snowflakeId >> shiftWorkerId) & maxWorkerId)
}

func GetSequenceFromId(snowflakeId uint64) uint16 {
	return uint16(snowflakeId & maxSequence)
}
