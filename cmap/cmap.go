package cmap

import (
	"errors"
)

// ConcurrentMap is a concurrent safely Map, support group shards
// every group has replicasNum shard
// replicaNum must be 2^n for ensure shard Index
// shardNum = groupNum * replicasNum
type ConcurrentMap[K comparable, V any] struct {
	shards     []*ConcurrentMapShared[K, V]
	hash       HashFunc[K]
	groupNum   uint32 // 可以不是2的整数次幂
	replicaNum uint32 // 必须是2的整数次幂
}

// 因为shardNum的数量是由应用程序决定，不一定是2^n,而replicaMask的数量则是固定的
func create[K comparable, V any](opts ...Option[K, V]) ConcurrentMap[K, V] {
	m := ConcurrentMap[K, V]{}
	for _, opt := range opts {
		opt(&m)
	}

	if m.replicaNum&(m.replicaNum-1) != 0 {
		panic("replicas must be pow(2,n)")
	}

	if m.hash == nil {
		panic("HashFunc can not be nil ")
	}

	if m.groupNum == 0 {
		m.groupNum = defaultGroupNum
	}

	if m.replicaNum == 0 {
		m.replicaNum = defaultReplicaNum
	}
	m.shards = make([]*ConcurrentMapShared[K, V], m.groupNum*m.replicaNum)
	for i := uint32(0); i < m.groupNum*m.replicaNum; i++ {
		m.shards[i] = &ConcurrentMapShared[K, V]{items: make(map[K]V)}
	}
	return m
}

func New[K string, V any](opt ...Option[K, V]) ConcurrentMap[K, V] {
	return create[K, V](opt...)
}
func NewDefault[V any]() ConcurrentMap[string, V] {
	return create[string, V](WithHashFunc[string, V](Fnv32[string]))
}

// GetShard returns shard under given key
func (m ConcurrentMap[K, V]) GetShard(key K) *ConcurrentMapShared[K, V] {
	return m.shards[m.calcIndex(key)]
}

// calcIndex 根据hash函数得到一个hash值，hash值和groupNum取余得到groupIndex，和replicas取余得到replicaIndex
// shardNum = groupNum * replicasNum
// shardIndex = groupIndex * replicasNum + replicaIndex
func (m ConcurrentMap[K, V]) calcIndex(key K) uint32 {
	shardVal := m.hash(key)
	return (shardVal%m.groupNum)*m.replicaNum + shardVal&(m.replicaNum-1)
}

func (m ConcurrentMap[K, V]) ClearShard(index uint32) error {
	if index > uint32(len(m.shards)) {
		return errors.New("index over flow shard num")
	}
	// 根据shard找到所有的index，然后清空
	start := m.replicaNum * index
	for i := uint32(0); i < m.replicaNum; i++ {
		m.shards[start+i].Lock()
		m.shards[start+i].items = make(map[K]V)
		m.shards[start+i].Unlock()
	}
	return nil
}

func (m ConcurrentMap[K, V]) MSet(data map[K]V) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

// Sets the given value under the specified key.
func (m ConcurrentMap[K, V]) Set(key K, value V) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// Callback to return new element to be inserted into the map
// It is called while lock is held, therefore it MUST NOT
// try to access other keys in same map, as it can lead to deadlock since
// Go sync.RWLock is not reentrant
type UpsertCb[V any] func(exist bool, valueInMap V, newValue V) V

// Insert or Update - updates existing element or inserts a new one using UpsertCb
func (m ConcurrentMap[K, V]) Upsert(key K, value V, cb UpsertCb[V]) (res V) {
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	res = cb(ok, v, value)
	shard.items[key] = res
	shard.Unlock()
	return res
}

// Sets the given value under the specified key if no value was associated with it.
func (m ConcurrentMap[K, V]) SetIfAbsent(key K, value V) bool {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	_, ok := shard.items[key]
	if !ok {
		shard.items[key] = value
	}
	shard.Unlock()
	return !ok
}

// Get retrieves an element from map under given key.
func (m ConcurrentMap[K, V]) Get(key K) (V, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Looks up an item under specified key
func (m ConcurrentMap[K, V]) Has(key K) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove removes an element from the map.
func (m ConcurrentMap[K, V]) Remove(key K) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// RemoveCb is a callback executed in a map.RemoveCb() call, while Lock is held
// If returns true, the element will be removed from the map
type RemoveCb[K any, V any] func(key K, v V, exists bool) bool

// RemoveCb locks the shard containing the key, retrieves its current value and calls the callback with those params
// If callback returns true and element exists, it will remove it from the map
// Returns the value returned by the callback (even if element was not present in the map)
func (m ConcurrentMap[K, V]) RemoveCb(key K, cb RemoveCb[K, V]) bool {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, ok := shard.items[key]
	remove := cb(key, v, ok)
	if remove && ok {
		delete(shard.items, key)
	}
	shard.Unlock()
	return remove
}

// Pop removes an element from the map and returns it
func (m ConcurrentMap[K, V]) Pop(key K) (v V, exists bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}

func (m ConcurrentMap[K, V]) Count() int {
	count := 0
	for i := 0; i < len(m.shards); i++ {
		shard := m.shards[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// IsEmpty checks if map is empty.
func (m ConcurrentMap[K, V]) IsEmpty() bool {
	return m.Count() == 0
}
