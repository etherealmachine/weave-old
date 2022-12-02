package main

import (
	"fmt"
	"log"
	"math/rand"
)

type Direction int

const (
	Below = Direction(0)
	Above = Direction(1)
	North = Direction(2)
	South = Direction(3)
	West  = Direction(4)
	East  = Direction(5)
)

func (d Direction) String() string {
	switch d {
	case Below:
		return "below"
	case Above:
		return "above"
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

var Neighbors = [6][3]int{
	{0, 0, -1}, // Below
	{0, 0, 1},  // Above
	{0, -1, 0}, // North
	{0, 1, 0},  // South
	{-1, 0, 0}, // West
	{1, 0, 0},  // East
}

type Generator struct {
	Domain               []*Tile
	Width, Depth, Height int
	Adj                  [][6][]int
	BanCount             *NDArray[int]
	Banned               *NDArray[bool]
	Support              *NDArray[int]
	Map                  *NDArray[int]
	Stack                [][4]int
	Seed                 int64
	RNG                  *rand.Rand
}

func NewGenerator(tiles Tilemap, w, h int, seed int64) *Generator {
	var domain []*Tile
	invDomain := make(map[string]map[int]int)
	var adj [][6][]int
	maxZ := 0
	for _, ys := range tiles {
		for _, zs := range ys {
			for z, tile := range zs {
				if z > maxZ {
					maxZ = z
				}
				exists := false
				for _, t := range domain {
					if t.Spritesheet == tile.Spritesheet && t.Index == tile.Index {
						exists = true
						break
					}
				}
				if !exists {
					if invDomain[tile.Spritesheet] == nil {
						invDomain[tile.Spritesheet] = make(map[int]int)
					}
					invDomain[tile.Spritesheet][tile.Index] = len(domain)
					domain = append(domain, tile)
					adj = append(adj, [6][]int{})
				}
			}
		}
	}
	for x, ys := range tiles {
		for y, zs := range ys {
			for z, tile := range zs {
				for dir, o := range Neighbors {
					n := tiles.At(x+o[0], y+o[1], z+o[2])
					if n == nil {
						continue
					}
					i := invDomain[tile.Spritesheet][tile.Index]
					j := invDomain[n.Spritesheet][n.Index]
					exists := false
					for _, a := range adj[i][dir] {
						if a == j {
							exists = true
							break
						}
					}
					if !exists {
						adj[i][dir] = append(adj[i][dir], j)
					}
				}
			}
		}
	}
	return &Generator{
		Domain: domain,
		Width:  w,
		Height: h,
		Depth:  maxZ + 1,
		Adj:    adj,
		Seed:   seed,
	}
}

func (g *Generator) Index(t *Tile) int {
	for i, o := range g.Domain {
		if o == t {
			return i
		}
	}
	return -1
}

func (g *Generator) Init() {
	g.BanCount = NewNDArray[int](g.Width, g.Height, g.Depth)
	g.Banned = NewNDArray[bool](g.Width, g.Height, g.Depth, len(g.Domain))
	g.initializeSupport()
	g.Map = NewNDArray[int](g.Width, g.Height, g.Depth)
	g.RNG = rand.New(rand.NewSource(g.Seed))
}

func (g *Generator) Done() bool {
	if len(g.Stack) > 0 {
		curr := g.Stack[len(g.Stack)-1]
		g.Stack = g.Stack[:len(g.Stack)-1]
		x, y, z, i := curr[0], curr[1], curr[2], curr[3]
		g.ban(x, y, z, i)
		return false
	}
	return g.collapse()
}

func (g *Generator) leastEntropy() []int {
	max := -1
	maxI := -1
	for i, c := range g.BanCount.Array() {
		if g.Map.Array()[i] == 0 && len(g.Domain)-c >= 1 && c > max {
			max = c
			maxI = i
		}
	}
	if maxI == -1 {
		return nil
	}
	return g.BanCount.Coords(maxI)
}

func (g *Generator) collapse() bool {
	loc := g.leastEntropy()
	if loc == nil {
		return true
	}
	x, y, z := loc[0], loc[1], loc[2]
	ticket := g.RNG.Intn(len(g.Domain) - g.BanCount.At(x, y, z))
	fmt.Printf("collapsing picked ticket %d from %d options\n", ticket, len(g.Domain)-g.BanCount.At(x, y, z))
	for i := range g.Domain {
		if g.Banned.At(x, y, z, i) {
			continue
		}
		if ticket == 0 {
			fmt.Printf("collapsing to %d (%s:%d) at (%d, %d, %d)\n", i, g.Domain[i].Spritesheet, g.Domain[i].Index, x, y, z)
			g.Map.Set(i+1, x, y, z)
			g.verify()
		} else {
			g.Stack = append(g.Stack, [4]int{x, y, z, i})
		}
		ticket--
	}
	fmt.Printf("collapse added %d to the stack\n", len(g.Stack))
	return false
}

func (g *Generator) ban(x, y, z, i int) {
	if g.Banned.At(x, y, z, i) {
		return
	}
	fmt.Printf("banning %d at (%d, %d, %d)\n", i, x, y, z)
	g.Banned.Set(true, x, y, z, i)
	g.BanCount.Set(g.BanCount.At(x, y, z)+1, x, y, z)
	// for each possible neighbor, remove this tile from support in the given direction
	for d, o := range Neighbors {
		for _, n := range g.Adj[i][d] {
			nx, ny, nz := x+o[0], y+o[1], z+o[2]
			if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height || nz < 0 || nz >= g.Depth {
				continue
			}
			if g.Map.At(nx, ny, nz) > 0 {
				fmt.Printf("skipping support change for (%d, %d, %d)\n", nx, ny, nz)
				continue
			}
			g.Support.Set(g.Support.At(nx, ny, nz, d, n)-1, nx, ny, nz, d, n)
			if g.Support.At(nx, ny, nz, d, n) == 0 {
				fmt.Printf("removed last support for %d (%s:%d) at (%d, %d, %d)\n", n, g.Domain[n].Spritesheet, g.Domain[n].Index, nx, ny, nz)
				g.Stack = append(g.Stack, [4]int{nx, ny, nz, n})
			}
		}
	}
}

func (g *Generator) Readout() [][][]*Tile {
	tiles := make([][][]*Tile, g.Width)
	for x := 0; x < g.Width; x++ {
		tiles[x] = make([][]*Tile, g.Height)
		for y := 0; y < g.Height; y++ {
			var stack []*Tile
			for z := 0; z < g.Depth; z++ {
				if i := g.Map.At(x, y, z) - 1; i != -1 {
					stack = append(stack, g.Domain[i])
				}
			}
			tiles[x][y] = stack
		}
	}
	return tiles
}

func (g *Generator) initializeSupport() {
	g.Stack = nil
	g.Support = NewNDArray[int](g.Width, g.Height, g.Depth, len(Neighbors), len(g.Domain))
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			for z := 0; z < g.Depth; z++ {
				for i := range g.Domain {
					for d, o := range Neighbors {
						for range g.Adj[i][d] {
							nx, ny, nz := x+o[0], y+o[1], z+o[2]
							if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height || nz < 0 || nz >= g.Depth {
								continue
							}
							// for each possible neighbor, add this tile as support in the given direction
							g.Support.Set(g.Support.At(nx, ny, nz, d, i)+1, nx, ny, nz, d, i)
						}
					}
				}
			}
		}
	}
	g.debug()
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			for z := 0; z < g.Depth; z++ {
				for i := range g.Domain {
					support := 0
					for d := range Neighbors {
						support += g.Support.At(x, y, z, d, i)
					}
					if support == 0 {
						fmt.Printf("no support in any direction for %d (%s:%d) at (%d, %d, %d)\n", i, g.Domain[i].Spritesheet, g.Domain[i].Index, x, y, z)
						g.Stack = append(g.Stack, [4]int{x, y, z, i})
					}
				}
			}
		}
	}
}

func (g *Generator) verify() {
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			for z := 0; z < g.Depth; z++ {
				g.verifyBanCount(x, y, z)
				for i := range g.Domain {
					if !g.Banned.At(x, y, z, i) {
						g.verifySupport(x, y, z, i)
					}
				}
				g.verifyPlacement(x, y, z)
			}
		}
	}
}

func (g *Generator) verifyBanCount(x, y, z int) {
	got := g.BanCount.At(x, y, z)
	want := 0
	for i := range g.Domain {
		if g.Banned.At(x, y, z, i) {
			want++
		}
	}
	if got != want {
		log.Fatalf("incorrect ban count at %d, %d, %d: got %d, want %d", x, y, z, got, want)
	}
}

func (g *Generator) verifySupport(x, y, z, i int) {
	/*
		for d, o := range Neighbors {
			nx, ny, nz := x+o[0], y+o[1], z+o[2]
			if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height || nz < 0 || nz >= g.Depth {
				continue
			}
			exists := false
			for _, adj := range g.Adj[i][d] {
				if adj == n {
					exists = true
					break
				}
			}
			if !exists {
				log.Fatalf("incorrect placement at %d, %d, %d: %d (%s:%d) cannot have %d (%s:%d) %s",
					x, y, z,
					i, g.Domain[i].Spritesheet, g.Domain[i].Index,
					n, g.Domain[n].Spritesheet, g.Domain[n].Index,
					Direction(d))
			}
		}
	*/
}

func (g *Generator) verifyPlacement(x, y, z int) {
	i := g.Map.At(x, y, z) - 1
	if i == -1 {
		return
	}
	for d, o := range Neighbors {
		nx, ny, nz := x+o[0], y+o[1], z+o[2]
		if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height || nz < 0 || nz >= g.Depth {
			continue
		}
		n := g.Map.At(nx, ny, nz) - 1
		if n == -1 {
			continue
		}
		exists := false
		for _, adj := range g.Adj[i][d] {
			if adj == n {
				exists = true
				break
			}
		}
		if !exists {
			log.Fatalf("incorrect placement at %d, %d, %d: %d (%s:%d) cannot have %d (%s:%d) %s",
				x, y, z,
				i, g.Domain[i].Spritesheet, g.Domain[i].Index,
				n, g.Domain[n].Spritesheet, g.Domain[n].Index,
				Direction(d))
		}
	}
}

func (g *Generator) debug() {
	for i, tile := range g.Domain {
		fmt.Printf("%d %s:%d\n", i, tile.Spritesheet, tile.Index)
	}
	for i := range g.Domain {
		for d := range Neighbors {
			fmt.Printf("support for %d %s: %d\n", i, Direction(d), g.Support.At(0, 0, 0, d, i))
		}
	}
}
