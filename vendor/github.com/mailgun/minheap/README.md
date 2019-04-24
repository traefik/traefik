[![Build Status](https://drone.io/github.com/mailgun/minheap/status.png)](https://drone.io/github.com/mailgun/minheap/latest)

minheap
=======

Slightly more user-friendly heap on top of containers/heap.

```go

import "github.com/mailgun/minheap"
	

func toEl(i int) interface{} {
	return &i
}

func fromEl(i interface{}) int {
	return *(i.(*int))
}

mh := minheap.NewMinHeap()

el := &minheap.Element{
   Value:    toEl(1),
   Priority: 5,
}

mh.PushEl(el)
mh.PeekEl()
mh.Len()
mh.PopEl()

```
