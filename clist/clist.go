package clist

import (
	"container/list"
)

type element[T any] struct {
	val  T
	next *element[T]
}

type LinkedList[T any] struct {
	head *element[T]
	cnt  int
}

func (l *LinkedList[T]) Add(t T) {
	l.cnt++
	if l.head == nil {
		l.head = &element[T]{t, nil}
	} else {
		l.head.next = &element[T]{t, nil}
	}
}

func (l *LinkedList[T]) Pop() (t T, exist bool) {
	list.New()
	if l.head == nil {
		exist = false
		return
	}
	t, l.head = l.head.val, l.head.next
	exist = true
	l.cnt--
	return
}

func (l *LinkedList[T]) Len() int {
	return l.cnt
}
