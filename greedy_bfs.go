package main

import (
	"math/rand"
)

/*
GreedyBFS: greedily expand around any fixed tiles
*/
type GreedyBFS struct {
	*Analysis
	queue         [][2]int
	result        *NDArray[*int]
	width, height int
	rng           *rand.Rand
}

func NewGreedyBFS(analysis *Analysis, width, height int, fixed Tilemap, seed int64) *GreedyBFS {
	g := &GreedyBFS{
		Analysis: analysis,
		result:   NewNDArray[*int](width, height),
		width:    width,
		height:   height,
		rng:      rand.New(rand.NewSource(seed)),
	}
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			if fixed[x][y] != nil {
				g.queue = append(g.queue, [2]int{x, y})
				i := g.DomainIndex[fixed[x][y].Hash()]
				g.result.Set(&i, x, y)
			}
		}
	}
	if len(g.queue) == 0 {
		g.queue = append(g.queue, [2]int{0, 0})
	}
	return g
}

func (g *GreedyBFS) Done() bool {
	if len(g.queue) == 0 {
		return true
	}
	curr := g.queue[0]
	g.queue = g.queue[1:]
	x, y := curr[0], curr[1]
	if g.result.At(x, y) == nil {
		banned := make([]bool, len(g.Domain))
		for d, o := range Neighbors {
			nx, ny := x+o[0], y+o[1]
			if nx < 0 || ny < 0 || nx >= g.width || ny >= g.height {
				continue
			}
			if n := g.result.At(nx, ny); n != nil {
				adj := g.Adj.At(*n, int(Direction(d).Inverse()))
				for i := range banned {
					if !adj[i] {
						banned[i] = true
					}
				}
			}
		}
		winner := g.Lottery(g.rng, func(i int) bool {
			return !banned[i]
		})
		if winner == -1 {
			return true
		}
		g.result.Set(&winner, x, y)
	}
	for _, o := range Neighbors {
		nx, ny := x+o[0], y+o[1]
		if nx < 0 || ny < 0 || nx >= g.width || ny >= g.height {
			continue
		}
		if g.result.At(nx, ny) == nil {
			g.queue = append(g.queue, [2]int{nx, ny})
		}
	}
	return false
}

func (g *GreedyBFS) Result() [][]Stack {
	shape := g.result.Shape()
	r := make([][]Stack, shape[0])
	for x := 0; x < shape[0]; x++ {
		r[x] = make([]Stack, shape[1])
		for y := 0; y < shape[1]; y++ {
			if i := g.result.At(x, y); i != nil {
				r[x][y] = g.Domain[*i]
			}
		}
	}
	return r
}
