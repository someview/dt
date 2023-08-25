package cqueue

import (
	"runtime"
	"sync/atomic"
)

type Queue[T any] struct {
	Head atomic.Pointer[Node[T]]
	Tail atomic.Pointer[Node[T]]
}

type Node[T any] struct {
	val  T
	next atomic.Pointer[Node[T]]
}

func newNode[T any](val T) *Node[T] {
	return &Node[T]{val: val}
}

func (q *Queue[T]) Enqueue(val T) {
	node := newNode[T](val)
	ptr := atomic.Pointer[Node[T]]{}
	ptr.Store(node)
	for {
		tailVal := q.Tail.Load()
		if q.Tail.CompareAndSwap(tailVal, node) {
			return
		}
		runtime.Gosched() // 让出执行权
	}
}

func (q *Queue[T]) Dequeue() T {
	for {
		head := q.Head.Load()
		if head == nil {
			var res T
			return res
		}
		nextVal := head.next.Load()
		if q.Head.CompareAndSwap(head, nextVal) {
			return head.val
		}
		runtime.Gosched() // 让出执行权
	}
}
