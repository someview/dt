package clist

import (
	. "github.com/fengyeall111/dt/iterator"
)

type CList[T any] interface {
	Front() T
	AddFront(ele T) // 将元素添加到列表的末尾。
	PopFront() T

	Back() T
	AddBack(ele T)
	PopBack() T

	Add(ele T) bool
	Remove(ele T) bool   // 从列表中删除指定元素的第一个匹配项。
	Contains(ele T) bool // 检查列表是否包含指定元素。
	Size() int           // 返回列表中的元素数。
	isEmpty() bool       // 检查列表是否为空。
	Clear()              // 清空列表中的所有元素。
	Iter() Iterator[T]   // 返回一个迭代器，用于遍历列表中的元素
}
