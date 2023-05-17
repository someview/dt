// Package lastevent
// 描述：这是一个最新事件的包装器
// 消费者只消费最新的事件，只有初始化或者消费者明确load的时候，才会尝试从新加载事件
// 生成者产生的事件最多只有一个最新的事件，尚未被消费的历史事件会被最新事件覆盖掉
package lastevent

import "sync"

type LastEvent[T any] struct {
	mu sync.RWMutex
	c  chan T
	v  *T // 长度固定为1
	ok bool
}

func (l *LastEvent[T]) Put(t T) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.ok {
		l.v = &t
	} else {
		select {
		case l.c <- t:
			l.ok = false
			l.v = nil // 清空旧的事件
		default:
			l.v = &t
		}
	}
}

func (l *LastEvent[T]) Get() chan T {
	return l.c
}

func (l *LastEvent[T]) Load() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.v != nil {
		select {
		case l.c <- *l.v:
			l.ok = false
			l.v = nil // 清空旧的事件
		default:
		}
	} else {
		l.ok = true
	}
}

func NewLastEvent[T any](ready bool) LastEvent[T] {
	return LastEvent[T]{
		c:  make(chan T, 1), // 用缓冲区长度为1的管道
		ok: ready,
	}
}
