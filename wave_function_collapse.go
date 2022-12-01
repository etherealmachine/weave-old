package main

import (
	"math/rand"
)

var Neighbors = [6][3]int{
	{0, 0, 1},  // Above
	{0, 0, -1}, // Below
	{0, 1, 0},  // North
	{0, -1, 0}, // South
	{1, 0, 0},  // East
	{-1, 0, 0}, // West
}

func inverse(dir int) int {
	if dir%2 == 0 {
		return dir + 1
	}
	return dir - 1
}

type Generator struct {
	Domain               []*Tile
	Width, Depth, Height int
	Adj                  [][6][]int
}

func NewGenerator(tiles Tilemap, x, y, w, h int) *Generator {
	var domain []*Tile
	invDomain := make(map[string]map[int]int)
	var adj [][6][]int
	d := 0
	for _, ys := range tiles {
		for _, zs := range ys {
			for z, tile := range zs {
				if z > d {
					d = z
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
						adj[j][dir] = append(adj[j][inverse(dir)], i)
					}
				}
			}
		}
	}
	return &Generator{
		Domain: domain,
		Width:  w,
		Height: h,
		Depth:  d,
		Adj:    adj,
	}
}

func (g *Generator) Generate() [][][]*Tile {
	banCount := NewNDArray[int](g.Width, g.Height, g.Depth)
	banned := NewNDArray[bool](g.Width, g.Height, g.Depth, len(g.Domain))
	support := NewNDArray[int](g.Width, g.Height, g.Depth, len(Neighbors), len(g.Domain))
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
							support.Set(x, y, z, d, i, support.At(x, y, z, d, i)+1)
						}
					}
				}
			}
		}
	}
	q := NewHeap(banCount.Array())
	for len(q) > 0 {
		i := q.Pop()
		loc := banned.Coords(i)
		x, y, z := loc[0], loc[1], loc[2]
		ticket := rand.Intn(len(g.Domain) - banCount.At(x, y, z))
		var stack [][4]int
		for i := range g.Domain {
			if banned.At(x, y, z, i) {
				continue
			}
			if ticket != 0 {
				stack = append(stack, [4]int{x, y, z, i})
			}
			ticket--
		}
		for len(stack) > 0 {
			curr := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			x, y, z, i := curr[0], curr[1], curr[2], curr[3]
			if banned.At(x, y, z, i) {
				continue
			}
			banned.Set(true, x, y, z, i)
			banCount.Set(banCount.At(x, y, z)+1, x, y, z)
			for d, adj := range g.Adj[i] {
				nd := inverse(d)
				nx, ny, nz := x+Neighbors[nd][0], y+Neighbors[nd][1], z+Neighbors[nd][2]
				if nx < 0 || nx >= g.Width || ny < 0 || ny >= g.Height || nz < 0 || nz >= g.Depth {
					continue
				}
				for _, j := range adj {
					support.Set(support.At(nx, ny, nz, d, j)-1, nx, ny, nz, d, j)
					if support.At(nx, ny, nz, d, j) == 0 {
						stack = append(stack, [4]int{nx, ny, nz, j})
					}
				}
			}
		}
	}
	tiles := make([][][]*Tile, g.Width)
	for x := 0; x < g.Width; x++ {
		tiles[x] = make([][]*Tile, g.Height)
		for y := 0; y < g.Height; y++ {
			var stack []*Tile
			for z := 0; z < g.Depth; z++ {
				if banCount.At(x, y, z) == len(g.Domain)-1 {
					for i := range g.Domain {
						if !banned.At(x, y, z, i) {
							stack = append(stack, g.Domain[i])
						}
					}
				}
			}
			tiles[x][y] = stack
		}
	}
	return tiles
}
