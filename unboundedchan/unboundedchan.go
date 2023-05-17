package UnboundedChanchan

import "sync"

// UnboundedChan is an implementation of an UnboundedChan buffer which does not use
// extra goroutines. This is typically used for passing updates from one entity
// to another within gRPC.
//
// All methods on this type are thread-safe and don't block on anything except
// the underlying mutex used for synchronization.
//
// UnboundedChan supports values of any type to be stored in it by using a channel
// of `interface{}`. This means that a call to Put() incurs an extra memory
// allocation, and also that users need a type assertion while reading. For
// performance critical code paths, using UnboundedChan is strongly discouraged and
// defining a new type specific implementation of this buffer is preferred. See
// internal/transport/transport.go for an example of this.
type UnboundedChan[T any] struct {
	c       chan T
	mu      sync.Mutex
	backlog *linkedList[T]
}

// NewUnboundedChan returns a new instance of UnboundedChan.
func NewUnboundedChan[T any]() *UnboundedChan[T] {
	return &UnboundedChan[T]{c: make(chan T, 1)}
}

// Put adds t to the UnboundedChan buffer.
func (b *UnboundedChan[T]) Put(t T) {
	b.mu.Lock()
	if b.backlog.IsEmpty() { // first put in case the buffer is full. This is the common case. 	b.c <- t } // send it off to{
		select {
		case b.c <- t:
			b.mu.Unlock()
			return
		default:
		}
	}
	b.backlog.Add(t)
	b.mu.Unlock()
}

// Load sends the earliest buffered data, if any, onto the read channel
// returned by Get(). Users are expected to call this every time they read a
// value from the read channel.
func (b *UnboundedChan[T]) Load() {
	b.mu.Lock()
	val, ok := b.backlog.Pop()
	if ok {
		select {
		case b.c <- val:
		default:
		}
	}
	b.mu.Unlock()
}

// Get returns a read channel on which values added to the buffer, via Put(),
// are sent on.
//
// Upon reading a value from this channel, users are expected to call Load() to
// send the next buffered value onto the channel if there is any.
func (b *UnboundedChan[T]) Get() <-chan T {
	return b.c
}
