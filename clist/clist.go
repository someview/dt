package clist

import (
	"sync/atomic"
)

type LockFreeList[T any] struct {
	head *Node[T]
	tail *Node[T]
}

type Node[T any] struct {
	item T
	key  int
	next *atomic.Value // 使用atomic.Value作为next字段
}

type Window[T any] struct {
	pred *Node[T]
	curr *Node[T]
}

func NewLockFreeList[T any]() *LockFreeList[T] {
	head := &Node[T]{key: int(^uint(0) >> 1)}
	tail := &Node[T]{key: int(^uint(0) >> 1)}
	head.next = &atomic.Value{}
	head.next.Store(tail)
	tail.next = &atomic.Value{}
	tail.next.Store(nil)
	return &LockFreeList[T]{head: head, tail: tail}
}

func (l *LockFreeList[T]) Add(item T) bool {
	key := item.(int)
	for {
		window := l.find(l.head, key)
		pred, curr := window.pred, window.curr
		if curr.key == key {
			return false
		} else {
			node := &Node[T]{item: item, key: key, next: &atomic.Value{}}
			node.next.Store(curr)
			if pred.next.CompareAndSwap(curr, node) {
				return true
			}
		}
	}
}

func (l *LockFreeList[T]) Remove(item T) bool {
	key := item.(int)
	var snip bool
	for {
		window := l.find(l.head, key)
		pred, curr := window.pred, window.curr
		if curr.key != key {
			return false
		} else {
			succ := curr.next.Load().(*Node[T])
			snip = curr.next.CompareAndSwap(succ, succ, false, true)
			if !snip {
				continue
			}
			pred.next.CompareAndSwap(curr, succ, false, false)
			return true
		}
	}
}

func (l *LockFreeList[T]) find(head *Node[T], key int) *Window[T] {
	var pred, curr, succ *Node[T]
	var marked bool
	var snip bool
retry:
	for {
		pred = head
		curr = pred.next.Load().(*Node[T])
		for {
			succ = curr.next.Load().(*Node[T])
			for marked = curr.next.Load() != nil; marked; marked = curr.next.Load() != nil {
				snip = pred.next.CompareAndSwap(curr, succ, false, false)
				if !snip {
					continue retry
				}
				curr = succ
				succ = curr.next.Load().(*Node[T])
			}
			if curr.key >= key {
				return &Window[T]{pred: pred, curr: curr}
			}
			pred = curr
			curr = succ
		}
	}
}
