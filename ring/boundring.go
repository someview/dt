package ring

type FixedRing[T any] struct {
	Ring[T]
	cap int
}

func (r *FixedRing[T]) Add(t T) {
	if r.Len() >= r.cap {
		r.Ring.Pop()
	}
	r.Ring.Add(t)
}

// if not value exists, this will return default T zero value
// you should not store T zero value in a FixedRing
func (r *FixedRing[T]) Pop() T {
	return r.Ring.Pop()
}

func NewFixedRing[T any](cap int) *FixedRing[T] {
	return &FixedRing[T]{
		Ring: Ring[T]{},
		cap:  cap,
	}
}
