package cmap

import "sync"

var defaultGroupNum = uint32(16)
var defaultReplicaNum = uint32(1)

type HashFunc[T comparable] func(key T) uint32 // 根据Hash寻找到Shard
// A "thread" safe string to anything map.
type ConcurrentMapShared[K comparable, V any] struct {
	items        map[K]V
	sync.RWMutex // Read Write mutex, guards access to internal map.
}

type Option[K comparable, V any] func(*ConcurrentMap[K, V])

func WithHashFunc[K comparable, V any](hash HashFunc[K]) Option[K, V] {
	return func(m *ConcurrentMap[K, V]) {
		m.hash = hash
	}
}

func Fnv32[T string](key T) uint32 {
	hash := uint32(2166136261)
	const prime32 = uint32(16777619)
	keyLength := len(key)
	for i := 0; i < keyLength; i++ {
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}

func WithGroupNum[K comparable, V any](groupNum uint32) Option[K, V] {
	return func(m *ConcurrentMap[K, V]) {
		m.groupNum = groupNum
	}
}

func WithReplicaNum[K comparable, V any](replicaNum uint32) Option[K, V] {
	return func(m *ConcurrentMap[K, V]) {
		m.replicaNum = replicaNum
	}
}
