package main

import (
	"math"
	"math/rand"
)

type WFC struct {
	*Analysis
	width, height int
	banCount      *NDArray[int]
	banned        *NDArray[bool]
	support       *NDArray[int]
	result        *NDArray[*int]
	stack         [][3]int
	seed          int64
	rng           *rand.Rand
	failed        bool
}

func NewWFC(analysis *Analysis, width, height int, fixed Tilemap, seed int64) *WFC {
	g := &WFC{Analysis: analysis}
	g.width = width
	g.height = height
	g.seed = seed
	g.banCount = NewNDArray[int](g.width, g.height)
	g.banned = NewNDArray[bool](g.width, g.height, len(g.Domain))
	g.initializeSupport()
	g.result = NewNDArray[*int](g.width, g.height)
	for x, ys := range fixed {
		for y, tiles := range ys {
			i := g.DomainIndex[tiles.Hash()]
			if i > 0 {
				g.result.Set(&i, x, y)
			}
			for j := range g.Domain {
				if j != i {
					g.stack = append(g.stack, [3]int{x, y, j})
				}
			}
		}
	}
	g.rng = rand.New(rand.NewSource(g.seed))
	return g
}

func (g *WFC) Done() bool {
	for len(g.stack) > 0 {
		curr := g.stack[len(g.stack)-1]
		g.stack = g.stack[:len(g.stack)-1]
		x, y, i := curr[0], curr[1], curr[2]
		g.ban(x, y, i)
	}
	if g.collapse() {
		return true
	}
	return false
}

func (g *WFC) Result() [][]Stack {
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

func (g *WFC) leastEntropy() (int, int) {
	minX, minY := -1, -1
	minEntropy := math.MaxFloat64
	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			if entropy := g.entropy(x, y); entropy < minEntropy {
				minEntropy = entropy
				minX, minY = x, y
			}
		}
	}
	return minX, minY
}

func (g *WFC) entropy(x, y int) float64 {
	if g.result.At(x, y) != nil {
		return math.MaxFloat64
	}
	var e float64
	for i, p := range g.Probabilities {
		if p > 0 && !g.banned.At(x, y, i) {
			e -= p * math.Log(p)
		}
	}
	return e
}

func (g *WFC) collapse() bool {
	x, y := g.leastEntropy()
	if x < 0 || y < 0 {
		return true
	}
	winner := g.Lottery(g.rng, func(i int) bool {
		return !g.banned.At(x, y, i)
	})
	if winner == -1 {
		return true
	}
	for i := 0; i < len(g.Domain); i++ {
		if i != winner {
			g.stack = append(g.stack, [3]int{x, y, i})
		}
	}
	g.result.Set(&winner, x, y)
	return false
}

func (g *WFC) ban(x, y, i int) {
	if g.banned.At(x, y, i) {
		return
	}
	g.banned.Set(true, x, y, i)
	banCount := g.banCount.At(x, y) + 1
	g.banCount.Set(banCount, x, y)
	if banCount == len(g.Domain) {
		g.failed = false
		return
	}
	// for each possible neighbor, remove this tile from support in the given direction
	for d, o := range Neighbors {
		nx, ny := x+o[0], y+o[1]
		if nx < 0 || nx >= g.width || ny < 0 || ny >= g.height {
			continue
		}
		if g.result.At(nx, ny) != nil {
			continue
		}
		for n := range g.Adj.At(i, d) {
			g.support.Set(g.support.At(nx, ny, n, d)-1, nx, ny, n, d)
			if g.support.At(nx, ny, n, d) == 0 {
				g.stack = append(g.stack, [3]int{nx, ny, n})
			}
		}
	}
}

func (g *WFC) initializeSupport() {
	g.stack = nil
	g.support = NewNDArray[int](g.width, g.height, len(g.Domain), len(Neighbors))
	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			for i := range g.Domain {
				for d := range Neighbors {
					support := len(g.Adj.At(i, int(Direction(d).Inverse())))
					if support == 0 {
						g.stack = append(g.stack, [3]int{x, y, i})
					} else {
						g.support.Set(support, x, y, i, d)
					}
				}
			}
		}
	}
}
