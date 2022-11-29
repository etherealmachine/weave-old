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

func (h Heap[E]) down(u int) bool {
	i := u
	for {
		j1 := 2*i + 1
		if j1 >= len(h) || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < len(h) && h[j2].Priority() < h[j1].Priority() {
			j = j2 // = 2*i + 2  // right child
		}
		if h[j].Priority() >= h[i].Priority() {
			break
		}
		h[i], h[j] = h[j], h[i]
		i = j
	}
	return i > u
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

func (h *Heap[E]) Fix(e E, i int) {
	if !h.down(i) {
		h.up(i)
	}
}
