package main

import (
	"fmt"
	"log"
	"math"
	"math/rand"
)

type Direction int

const (
	North = Direction(0)
	South = Direction(1)
	West  = Direction(2)
	East  = Direction(3)
)

func (d Direction) String() string {
	switch d {
	case North:
		return "north"
	case South:
		return "south"
	case West:
		return "west"
	case East:
		return "east"
	default:
		return "unknown"
	}
}

func (d Direction) Inverse() Direction {
	if d%2 == 0 {
		return d + 1
	}
	return d - 1
}

var Neighbors = [4][2]int{
	{0, -1}, // North
	{0, 1},  // South
	{-1, 0}, // West
	{1, 0},  // East
}

type Generator struct {
	Domain        []Stack
	Probabilities []float64
	Width, Height int
	Adj           *NDArray[map[int]bool]
	BanCount      *NDArray[int]
	Banned        *NDArray[bool]
	Support       *NDArray[int]
	Map           *NDArray[*int]
	Stack         [][3]int
	Seed          int64
	RNG           *rand.Rand
	Verify        bool
}

func NewGenerator(tilemap Tilemap, w, h int, seed int64) *Generator {
	type e struct {
		Index int
		Stack Stack
	}
	domainMap := map[string]*e{
		"": {Index: 0, Stack: nil},
	}
	for _, ys := range tilemap {
		for _, tiles := range ys {
			if h := tiles.Hash(); domainMap[h] == nil {
				domainMap[h] = &e{
					Index: len(domainMap),
					Stack: tiles,
				}
			}
		}
	}
	probs := make([]float64, len(domainMap))
	adj := NewNDArray[map[int]bool](len(domainMap), len(Neighbors))
	for x, ys := range tilemap {
		for y, tiles := range ys {
			i := domainMap[tiles.Hash()]
			probs[i.Index]++
			for d, o := range Neighbors {
				nx, ny := x+o[0], y+o[1]
				n := domainMap[tilemap[nx][ny].Hash()]
				a := adj.At(i.Index, d)
				if a == nil {
					a = make(map[int]bool)
					adj.Set(a, i.Index, d)
				}
				a[n.Index] = true
				di := int(Direction(d).Inverse())
				a = adj.At(n.Index, di)
				if a == nil {
					a = make(map[int]bool)
					adj.Set(a, n.Index, di)
				}
				a[i.Index] = true
			}
		}
	}
	var sum float64
	for _, count := range probs {
		sum += count
	}
	for i, count := range probs {
		probs[i] = count / sum
	}
	domain := make([]Stack, len(domainMap))
	for _, e := range domainMap {
		domain[e.Index] = e.Stack
	}
	return &Generator{
		Domain:        domain,
		Probabilities: probs,
		Width:         w,
		Height:        h,
		Adj:           adj,
		Seed:          seed,
	}
}

func (g *Generator) Init() {
	g.BanCount = NewNDArray[int](g.Width, g.Height)
	g.Banned = NewNDArray[bool](g.Width, g.Height, len(g.Domain))
	g.initializeSupport()
	g.Map = NewNDArray[*int](g.Width, g.Height)
	g.RNG = rand.New(rand.NewSource(g.Seed))
	if g.Verify {
		g.debugDomain()
	}
}

func (g *Generator) Done() bool {
	if g.Verify {
		g.verify()
		defer g.verify()
	}
	if g.collapse() {
		return true
	}
	for len(g.Stack) > 0 {
		curr := g.Stack[len(g.Stack)-1]
		g.Stack = g.Stack[:len(g.Stack)-1]
		x, y, i := curr[0], curr[1], curr[2]
		g.ban(x, y, i)
	}
	return false
}

func (g *Generator) leastEntropy() (int, int) {
	minX, minY := -1, -1
	minEntropy := math.MaxFloat64
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			if entropy := g.entropy(x, y); entropy < minEntropy {
				minEntropy = entropy
				minX, minY = x, y
			}
		}
	}
	return minX, minY
}

func (g *Generator) entropy(x, y int) float64 {
	if g.Map.At(x, y) != nil {
		return math.MaxFloat64
	}
	var e float64
	for i, p := range g.Probabilities {
		if p > 0 && !g.Banned.At(x, y, i) {
			e -= p * math.Log(p)
		}
	}
	return e
}

func (g *Generator) collapse() bool {
	x, y := g.leastEntropy()
	if x < 0 || y < 0 {
		return true
	}
	var ticketCount float64
	tickets := make(map[int]float64)
	for i := range g.Domain {
		if g.Banned.At(x, y, i) {
			continue
		}
		tickets[i] = g.Probabilities[i]
		ticketCount += g.Probabilities[i]
	}
	ticket := g.RNG.Float64() * ticketCount
	winner := -1
	for i := 0; i < len(g.Domain); i++ {
		ticket -= tickets[i]
		if winner == -1 && ticket <= 0 {
			winner = i
		} else {
			g.Stack = append(g.Stack, [3]int{x, y, i})
		}
	}
	if winner == -1 {
		return true
	}
	g.Map.Set(&winner, x, y)
	return false
}

func (g *Generator) ban(x, y, i int) {
	if g.Banned.At(x, y, i) {
		return
	}
	g.Banned.Set(true, x, y, i)
	g.BanCount.Set(g.BanCount.At(x, y)+1, x, y)
	// for each possible neighbor, remove this tile from support in the given direction
	for d, o := range Neighbors {
		nx, ny := x+o[0], y+o[1]
		if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height {
			continue
		}
		if g.Map.At(nx, ny) != nil {
			continue
		}
		for n := range g.Adj.At(i, d) {
			g.Support.Set(g.Support.At(nx, ny, n, d)-1, nx, ny, n, d)
			if g.Support.At(nx, ny, n, d) == 0 {
				g.Stack = append(g.Stack, [3]int{nx, ny, n})
			}
		}
	}
}

func (g *Generator) initializeSupport() {
	g.Stack = nil
	g.Support = NewNDArray[int](g.Width, g.Height, len(g.Domain), len(Neighbors))
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			for i := range g.Domain {
				for d := range Neighbors {
					support := len(g.Adj.At(i, int(Direction(d).Inverse())))
					if support == 0 {
						g.Stack = append(g.Stack, [3]int{x, y, i})
					} else {
						g.Support.Set(support, x, y, i, d)
					}
				}
			}
		}
	}
}

func (g *Generator) verify() {
	g.debug()
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			g.verifyBanCount(x, y)
			if g.Map.At(x, y) == nil {
				for i := range g.Domain {
					if !g.Banned.At(x, y, i) {
						g.verifySupport(x, y, i)
					}
				}
			}
			g.verifyPlacement(x, y)
		}
	}
}

func (g *Generator) verifyBanCount(x, y int) {
	got := g.BanCount.At(x, y)
	want := 0
	for i := range g.Domain {
		if g.Banned.At(x, y, i) {
			want++
		}
	}
	if got != want {
		log.Fatalf("incorrect ban count at (%d, %d): got %d, want %d", x, y, got, want)
	}
}

func (g *Generator) verifySupport(x, y, i int) {
}

func (g *Generator) verifyPlacement(x, y int) {
	i := g.Map.At(x, y)
	if i == nil {
		return
	}
	for d, o := range Neighbors {
		nx, ny := x+o[0], y+o[1]
		if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height {
			continue
		}
		n := g.Map.At(nx, ny)
		if n == nil {
			continue
		}
		if !g.Adj.At(*i, d)[*n] {
			log.Fatalf("incorrect placement at (%d, %d): %d cannot have %d %s",
				x, y,
				*i, *n,
				Direction(d))
		}
	}
}

func (g *Generator) debugDomain() {
	fmt.Print("domain:\n")
	for i, stack := range g.Domain {
		fmt.Printf("%d %s\n", i, stack.Hash())
	}
	fmt.Print("adjacency:\n")
	for i := range g.Domain {
		for d := range Neighbors {
			fmt.Printf("%d %s: ", i, Direction(d))
			for adj := range g.Adj.At(i, d) {
				fmt.Printf("%d ", adj)
			}
			fmt.Println()
		}
	}
}

func (g *Generator) debug() {
	fmt.Print("map:\n")
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			if i := g.Map.At(x, y); i != nil {
				fmt.Printf("%d", *i)
			} else {
				fmt.Print(" ")
			}
			if x+1 < g.Width {
				fmt.Print(", ")
			}
		}
		fmt.Println()
	}
	fmt.Print("banCount:\n")
	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			fmt.Printf("%d", g.BanCount.At(x, y))
			if x+1 < g.Width {
				fmt.Print(", ")
			}
		}
		fmt.Println()
	}
}
