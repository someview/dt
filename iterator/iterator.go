package iterator

type Iterator[T any] interface {
	Next() T
	HasNext() bool
}
