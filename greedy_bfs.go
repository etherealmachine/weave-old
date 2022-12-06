package main

import "log"

/*
GreedyBFS: greedily expand around any fixed tiles
*/
type GreedyBFS struct {
	*Analysis
	queue  [][2]int
	result *NDArray[*int]
}

func NewGreedyBFS(analysis *Analysis, width, height int, fixed Tilemap, seed int64) *GreedyBFS {
	g := &GreedyBFS{Analysis: analysis, result: NewNDArray[*int](width, height)}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if fixed[x][y] != nil {
				g.queue = append(g.queue, [2]int{x, y})
				i := g.DomainIndex[fixed[x][y].Hash()]
				g.result.Set(&i, x, y)
			}
		}
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if fixed[x][y] == nil {
				g.queue = append(g.queue, [2]int{x, y})
			}
		}
	}
	return g
}

func (g *GreedyBFS) Done() bool {
	if len(g.queue) == 0 {
		return true
	}
	curr := g.queue[0]
	g.queue = g.queue[1:]
	if g.result.At(curr[0], curr[1]) == nil {
		return false
	}
	log.Println(g.Domain[*g.result.At(curr[0], curr[1])])
	return false
}

func (g *GreedyBFS) Result() [][]Stack {
	return nil
}
