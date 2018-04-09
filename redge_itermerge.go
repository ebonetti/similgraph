package similgraph

import (
	"container/heap"
)

type redgeIterMerge struct{ h redgeIterHeap }

func redgeIterMergeFrom(nexts ...func() (info rEdge, ok bool)) *redgeIterMerge {
	h := make(redgeIterHeap, 0, len(nexts))
	for _, next := range nexts {
		if info, ok := next(); ok {
			h = append(h, redgeIterator{info, next})
		}
	}
	heap.Init(&h)
	return &redgeIterMerge{h}
}
func (m *redgeIterMerge) Push(next func() (rEdge, bool)) {
	heap.Push(&m.h, next)
}
func (m redgeIterMerge) Peek() (info rEdge, ok bool) {
	if len(m.h) > 0 {
		info, ok = m.h[0].Info, true
	}
	return
}
func (m *redgeIterMerge) Next() (info rEdge, ok bool) {
	h := m.h
	if len(h) == 0 {
		return
	}
	info, ook := h[0].Next()
	h[0].Info, info = info, h[0].Info
	if ook {
		heap.Fix(&m.h, 0)
	} else {
		heap.Pop(&m.h)
	}
	return info, true
}

type redgeIterHeap []redgeIterator
type redgeIterator struct {
	Info rEdge
	Next func() (info rEdge, ok bool)
}

func (h redgeIterHeap) Len() int {
	return len(h)
}
func (h redgeIterHeap) Less(i, j int) bool {
	return h[i].Info.Less(h[j].Info)
}
func (h redgeIterHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}
func (h *redgeIterHeap) Push(x interface{}) {
	next := x.(func() (rEdge, bool))
	if info, ok := next(); ok {
		*h = append(*h, redgeIterator{info, next})
	}
}
func (h *redgeIterHeap) Pop() interface{} {
	_h := *h
	n := len(_h)
	x := _h[n-1]
	_h[n-1] = redgeIterator{}
	*h = _h[0 : n-1]
	return x
}
