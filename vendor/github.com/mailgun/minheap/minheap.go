package minheap

import (
	"container/heap"
)

// An Element is something we manage in a priority queue.
type Element struct {
	Value    interface{}
	Priority int // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type MinHeap []*Element

func NewMinHeap() *MinHeap {
	mh := &MinHeap{}
	heap.Init(mh)
	return mh
}

func (mh MinHeap) Len() int { return len(mh) }

func (mh MinHeap) Less(i, j int) bool {
	return mh[i].Priority < mh[j].Priority
}

func (mh MinHeap) Swap(i, j int) {
	mh[i], mh[j] = mh[j], mh[i]
	mh[i].index = i
	mh[j].index = j
}

func (mh *MinHeap) Push(x interface{}) {
	n := len(*mh)
	item := x.(*Element)
	item.index = n
	*mh = append(*mh, item)
}

func (mh *MinHeap) Pop() interface{} {
	old := *mh
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*mh = old[0 : n-1]
	return item
}

func (mh *MinHeap) PushEl(el *Element) {
	heap.Push(mh, el)
}

func (mh *MinHeap) PopEl() *Element {
	el := heap.Pop(mh)
	return el.(*Element)
}

func (mh *MinHeap) PeekEl() *Element {
	items := *mh
	return items[0]
}

// update modifies the priority and value of an Item in the queue.
func (mh *MinHeap) UpdateEl(el *Element, priority int) {
	heap.Remove(mh, el.index)
	el.Priority = priority
	heap.Push(mh, el)
}

func (mh *MinHeap) RemoveEl(el *Element) {
	heap.Remove(mh, el.index)
}
