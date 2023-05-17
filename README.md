# ext-dt
extended datastructure
## orderedmap
- 双向链表+map构成的map，可以有序迭代

## concurrent-map
- ConcurrentMap is a concurrent safely Map, support group shards
- every group has replicasNum shard
- replicaNum must be 2^n for ensure shard Index

## ring
- 双向环形可覆盖缓冲区

## lastevent
- 这是一个最新事件的包装器
- 消费者只消费最新的事件，只有初始化或者消费者明确load的时候，才会尝试从新加载事件
- 生成者产生的事件最多只有一个最新的事件，尚未被消费的历史事件会被最新事件覆盖掉
