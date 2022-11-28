package main

type Prioritizable interface {
	Priority() int
}

type Heap[E Prioritizable] []E

func (h Heap[E]) Init() {
	for i := (len(h) - 1) / 2; i >= 0; i-- {
		h.down(i)
	}
}

func (h Heap[E]) Len() int {
	return len(h)
}

func (h Heap[E]) down(u int) {
	v := u
	if 2*u+1 < len(h) && h[2*u+1].Priority() < h[v].Priority() {
		v = 2*u + 1
	}
	if 2*u+2 < len(h) && h[2*u+2].Priority() < h[v].Priority() {
		v = 2*u + 2
	}
	if v != u {
		h[v], h[u] = h[u], h[v]
		h.down(v)
	}
}

func (h Heap[E]) up(u int) {
	for u != 0 && h[(u-1)/2].Priority() > h[u].Priority() {
		h[(u-1)/2], h[u] = h[u], h[(u-1)/2]
		u = (u - 1) / 2
	}
}

func (h *Heap[E]) Push(e E) {
	*h = append(*h, e)
	h.up(len(*h) - 1)
}

func (h *Heap[E]) Pop() E {
	x := (*h)[0]
	n := len(*h)
	(*h)[0], (*h)[n-1] = (*h)[n-1], (*h)[0]
	*h = (*h)[:n-1]
	h.down(0)
	return x
}
