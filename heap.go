package main

type Heap [][2]int

func NewHeap(priorities []int) Heap {
	h := Heap(make([][2]int, len(priorities)))
	for i, p := range priorities {
		h[i] = [2]int{i, p}
	}
	for i := (len(h) - 1) / 2; i >= 0; i-- {
		h.down(i)
	}
	return h
}

func (h Heap) down(u int) bool {
	i := u
	for {
		j1 := 2*i + 1
		if j1 >= len(h) || j1 < 0 { // j1 < 0 after int overflow
			break
		}
		j := j1 // left child
		if j2 := j1 + 1; j2 < len(h) && h[j2][1] < h[j1][1] {
			j = j2 // = 2*i + 2  // right child
		}
		if h[j][1] >= h[i][1] {
			break
		}
		h[i], h[j] = h[j], h[i]
		i = j
	}
	return i > u
}

func (h Heap) up(u int) {
	for u != 0 && h[(u-1)/2][1] > h[u][1] {
		h[(u-1)/2], h[u] = h[u], h[(u-1)/2]
		u = (u - 1) / 2
	}
}

func (h *Heap) Pop() int {
	x := (*h)[0]
	n := len(*h)
	(*h)[0], (*h)[n-1] = (*h)[n-1], (*h)[0]
	*h = (*h)[:n-1]
	h.down(0)
	return x[0]
}

func (h *Heap) Update(e int) {
	//if !h.down(i) {
	//	h.up(i)
	//}
}
