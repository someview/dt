// offer ring and fixedring, not concurrent safely
// you should not store T zero value in ring,because pop() return default T zero value
package ring

type Node[T any] struct {
	val  T
	next *Node[T]
}

// Ring doubleLinkList
type Ring[T any] struct { // 并发不安全，需要保证准确结果的情况下需要保证Add和Pop的调用不会并发
	head *Node[T]
	tail *Node[T]
	len  int
}

func (l *Ring[T]) Add(t T) {
	newNode := &Node[T]{
		val: t,
	}
	if l.head == nil {
		l.head = newNode
		l.tail = newNode
		newNode.next = l.head
	} else {
		l.tail.next = newNode
		l.tail = newNode
		newNode.next = l.head
	}
	l.len++
}

// if not value exists, this will return default T zero value
// you should not store T zero value in a FixedRing
func (l *Ring[T]) Pop() T {
	if l.head == nil {
		l.len = 0
		var t T
		return t
	}
	val := l.head.val
	if l.len == 1 {
		l.head = nil
		l.tail = nil
	} else {
		l.head = l.head.next
		l.tail.next = l.head
	}
	l.len--
	return val
}

func (l *Ring[T]) Len() int {
	return l.len
}
