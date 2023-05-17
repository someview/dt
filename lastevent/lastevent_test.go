package lastevent

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvent_Load(t *testing.T) {
	event := NewLastEvent[int](true)
	// 首先设置event
	event.Put(1)
	event.Put(2)
	event.Put(3)

	resultChan := event.Get()
	val := <-resultChan
	assert.Equal(t, 1, val) // 第一次获取的值是1
	event.Put(4)
	select {
	case <-resultChan:
		assert.Error(t, fmt.Errorf("在消费者未通知生产者之前,仍然接收到数据"))
	default:
		t.Log("在消费者未通知生产者之前,接收不到数据")
	}
	event.Load() // 加载缓存数据,同时允许新的数据进行put
	val = <-resultChan
	assert.Equal(t, 4, val) // 第二次获取的值是cpu
}
func TestEvent_LoadCacheValue(t *testing.T) {
	event := NewLastEvent[int](true)
	// 首先设置event
	event.Put(1)
	event.Put(2)
	event.Put(3)

	resultChan := event.Get()
	val := <-resultChan
	assert.Equal(t, 1, val) // 第一次获取的值是1
	select {
	case <-resultChan:
		assert.Error(t, fmt.Errorf("在消费者未通知生产者之前,仍然接收到数据"))
	default:
		t.Log("在消费者未通知生产者之前,接收不到数据")
	}
	event.Load() // 加载缓存数据,同时允许新的数据进行put
	val = <-resultChan
	assert.Equal(t, 3, val) // 第二次获取的值是cpu

	event.Load() // 加载缓存数据,同时允许新的数据进行put
	select {
	case <-resultChan:
		t.Fatal("无缓存, load之后不应该收到数据")
	default:
		t.Log("无缓存，本身不会接收到数据")
	}
}
