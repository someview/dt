package UnboundedChanchan

type element[T any] struct {
	val  T
	next *element[T]
}

type linkedList[T any] struct {
	head *element[T]
}

func (l *linkedList[T]) Add(t T) {
	if l.head == nil {
		l.head = &element[T]{t, nil}
	} else {
		l.head.next = &element[T]{t, nil}
	}
}

func (l *linkedList[T]) Pop() (t T, exist bool) {
	if l.head == nil {
		exist = false
		return
	}
	t, l.head = l.head.val, l.head.next
	exist = true
	return
}

func (l *linkedList[T]) IsEmpty() bool {
	return l.head == nil
}
