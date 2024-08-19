package pool

import "sync"

type Pool[T any] interface {
	Get() T
	Put(T)
}

type PrePut[T any] func(T)
type PostGet[T any] func(T)

type Generator func() any
type PoolOption[T any] func(opt *option[T])

func WithPrePut[T any](prePut PrePut[T]) PoolOption[T] {
	return func(opt *option[T]) {
		opt.PrePut = prePut
	}
}

func WithPostGet[T any](postGet PostGet[T]) PoolOption[T] {
	return func(opt *option[T]) {
		opt.PostGet = postGet
	}
}

type option[T any] struct {
	PrePut[T]
	PostGet[T]
}

func newDefaultOption[T any]() *option[T] {
	return &option[T]{}
}

type syncPool[T any] struct {
	pool sync.Pool
	option[T]
}

func (s *syncPool[T]) Get() T {
	t := s.pool.Get().(T)
	if s.PostGet != nil {
		s.PostGet(t)
	}
	return t
}

func (s *syncPool[T]) Put(t T) {
	if s.PrePut != nil {
		s.PrePut(t)
	}
	s.pool.Put(t)
}

func NewSyncPool[T any](gen Generator, opts ...PoolOption[T]) Pool[T] {
	dopts := newDefaultOption[T]()
	for _, opt := range opts {
		opt(dopts)
	}
	return &syncPool[T]{
		pool:   sync.Pool{New: gen},
		option: *dopts,
	}
}
