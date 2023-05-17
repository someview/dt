package pipe

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

type Chan[T any] chan T

type Buffer interface {
	Cap() int
	Len() int
}

type Piper interface {
	Buffer
	InTime() error
	OutTime() error
	GetAvgTime() float64
	GetUsageRate() float64 // 当前的管道使用率
}

type Pipe[T any] struct {
	Chan[T]
	ignoreCount int64
	inTimes     timeSave
	subTime     int64
	avgTime     float64
	subPointer  int
	size        int // 多少条数据统计一次
	readFlag    atomic.Bool
	writeLock   sync.Mutex
	writeFlag   atomic.Bool
}

func NewPipe[T any](cap int, num int) Pipe[T] { // 连续采样条数
	res := Pipe[T]{
		Chan: make(Chan[T], cap),
		size: num,
	}
	return res
}

type timeSave struct {
	time []int64
	p    int // 当前指针
	size int // 切片大小
}

func (p *Pipe[T]) InTime() error {
	if !p.writeFlag.Load() { // 这儿不用继续统计
		return nil
	}
	p.writeLock.Lock()
	defer p.writeLock.Unlock()
	t := time.Now().UnixMilli()
	p.inTimes.time[p.inTimes.p] = t
	p.inTimes.p += 1
	if p.inTimes.p >= p.size {
		p.writeFlag.Store(false)
		p.inTimes.p = 0 // 将指针设置为0
	}
	return nil
}

// OutTime() 这种采样方式的弊端在于，即使promethues只能采集一次数据，这儿也需要频繁的采样，所以更好的方案应该是每隔多少条消息采样一条消息
func (p *Pipe[T]) OutTime() error { // 确保outTime是单协程调用
	if p.ignoreCount > 0 {
		p.ignoreCount -= 1
		return errors.New("同步中")
	}
	p.subTime += time.Now().UnixMilli() - p.inTimes.time[p.subPointer]
	p.subPointer += 1
	if p.subPointer >= p.size { //
		p.subPointer = 0
		p.ignoreCount = int64(len(p.Chan))
		p.avgTime = float64(p.subTime*1e5/int64(p.size)) / 1e5
		p.writeFlag.Store(true)
	}
	return nil
}

func (p *Pipe[T]) GetAvgTime() float64 { // 设置标志位,已经读取过了
	return p.avgTime
}

func (p *Pipe[T]) GetUsageRate() float64 {
	return float64(len(p.Chan)*100000/cap(p.Chan)) / 100000
}

func (p *Pipe[T]) Len() int { // 设置标志位,已经读取过了
	return len(p.Chan)
}

func (p *Pipe[T]) Cap() int {
	return cap(p.Chan)
}
