package main

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
)

type Tileset map[string]map[int]bool

func NewTileset(tiles ...*Tile) Tileset {
	s := make(Tileset)
	for _, t := range tiles {
		s.Add(t)
	}
	return s
}

func (s Tileset) Add(tiles ...*Tile) int {
	count := 0
	for _, t := range tiles {
		if s[t.Spritesheet] == nil {
			s[t.Spritesheet] = make(map[int]bool)
		}
		if !s[t.Spritesheet][t.Index] {
			count++
		}
		s[t.Spritesheet][t.Index] = true
	}
	return count
}

func (s Tileset) Remove(tiles ...*Tile) int {
	count := 0
	for _, t := range tiles {
		if s[t.Spritesheet] == nil {
			s[t.Spritesheet] = make(map[int]bool)
		}
		if s[t.Spritesheet][t.Index] {
			count++
		}
		s[t.Spritesheet][t.Index] = false
	}
	return count
}

// Union adds all the tiles in other to the tileset
// it returns true if any tiles were added
func (s Tileset) Union(other Tileset) bool {
	diff := false
	for n, tiles := range other {
		for i, exists := range tiles {
			if exists {
				if s[n] == nil {
					s[n] = make(map[int]bool)
				}
				if !s[n][i] {
					diff = true
				}
				s[n][i] = true
			}
		}
	}
	return diff
}

// Intersect keeps tiles in s that are also in other
// it returns true if the result is non-empty
func (s Tileset) Intersect(other Tileset) bool {
	empty := true
	log.Printf("%v | %v", s, other)
	for n, tiles := range s {
		for i, exists := range tiles {
			if exists {
				if other[n] == nil || !other[n][i] {
					s[n][i] = false
				} else {
					empty = false
				}
			}
		}
	}
	return empty
}

func (s Tileset) Tiles() []*Tile {
	var tiles []*Tile
	for spritesheet, indices := range s {
		for index, exists := range indices {
			if exists {
				tiles = append(tiles, &Tile{
					Spritesheet: spritesheet,
					Index:       index,
				})
			}
		}
	}
	return tiles
}

// tileset name -> tile index -> Direction -> adjacent tiles
type AdjacencySet map[string]map[int]map[Direction]Tileset

func (s AdjacencySet) At(spritesheet string, index int, direction Direction) Tileset {
	return s[spritesheet][index][direction]
}

type Generator struct {
	Adj AdjacencySet
	Pos Possibilities
}

func NewGenerator(m *Tilemap, x, y, w, h int) *Generator {
	adj := make(AdjacencySet)
	d := 0
	firstLayer := NewTileset()
	for x, ys := range m.Tiles {
		for y, tiles := range ys {
			for z, tile := range tiles {
				if z > d {
					d = z
				}
				if z == 0 {
					firstLayer.Add(tile)
				}
				for dir, offset := range neighborOffsets {
					addEdge(adj, m.TileAt(x+offset[0], y+offset[1], z+offset[2]), dir, tile)
				}
			}
		}
	}
	firstLayerTiles := firstLayer.Tiles()
	pos := make(Possibilities, w)
	for x := 0; x < w; x++ {
		pos[x] = make([][]PossibilitySet, h)
		for y := 0; y < h; y++ {
			pos[x][y] = make([]PossibilitySet, d)
			pos[x][y][0] = make(PossibilitySet, len(firstLayerTiles))
			for i, tile := range firstLayerTiles {
				pos[x][y][0][i] = &Possibility{
					X:    x,
					Y:    y,
					Z:    0,
					Tile: tile,
				}
			}
		}
	}
	return &Generator{
		Adj: adj,
		Pos: pos,
	}
}

type Direction string

const (
	Above = Direction("above")
	Below = Direction("below")
	North = Direction("north")
	South = Direction("south")
	East  = Direction("east")
	West  = Direction("west")
)

func (d Direction) Inverse() Direction {
	switch d {
	case Above:
		return Below
	case Below:
		return Above
	case North:
		return South
	case South:
		return North
	case East:
		return West
	case West:
		return East
	}
	return ""
}

var neighborOffsets = map[Direction][3]int{
	Above: {0, 0, 1},
	Below: {0, 0, -1},
	North: {0, 1, 0},
	South: {0, -1, 0},
	East:  {1, 0, 0},
	West:  {-1, 0, 0},
}

func addEdge(adj AdjacencySet, t1 *Tile, direction Direction, t2 *Tile) {
	if t1 == nil && t2 == nil {
		return
	}
	if t1 == nil {
		t1 = &Tile{
			Spritesheet: "empty",
			Index:       0,
		}
	}
	if t2 == nil {
		t2 = &Tile{
			Spritesheet: "empty",
			Index:       0,
		}
	}
	if adj[t1.Spritesheet] == nil {
		adj[t1.Spritesheet] = make(map[int]map[Direction]Tileset)
	}
	if adj[t1.Spritesheet][t1.Index] == nil {
		adj[t1.Spritesheet][t1.Index] = make(map[Direction]Tileset)
	}
	if adj[t1.Spritesheet][t1.Index][direction] == nil {
		adj[t1.Spritesheet][t1.Index][direction] = NewTileset()
	}
	if adj[t1.Spritesheet][t1.Index][direction].Add(t1) > 0 {
		addEdge(adj, t2, direction.Inverse(), t1)
	}
}

type Possibility struct {
	X, Y, Z   int
	Tile      *Tile
	Neighbors [6]*Possibility
}

type PossibilitySet []*Possibility

func (s PossibilitySet) Priority() int {
	return len(s)
}

// x, y, z -> possible tiles in that position
type Possibilities [][][]PossibilitySet

func (s Possibilities) At(x, y, z int) PossibilitySet {
	if x < 0 || x >= len(s) {
		return nil
	}
	if y < 0 || y >= len(s[x]) {
		return nil
	}
	if z < 0 || z >= len(s[x][y]) {
		return nil
	}
	return s[x][y][z]
}

func (g *Generator) Generate() [][][]*Tile {
	var q Heap[PossibilitySet]
	q.Init()
	for _, ys := range g.Pos {
		for _, zs := range ys {
			for _, possible := range zs {
				q.Push(possible)
			}
		}
	}
	for q.Len() > 0 {
		set := q.Pop()
		if len(set) == 0 {
			continue
		}
		choice := set[rand.Intn(len(set))]
		g.Pos[choice.X][choice.Y][choice.Z] = PossibilitySet{choice}
		// propagate changes
	}
	tiles := make([][][]*Tile, len(g.Pos))
	for x, ys := range g.Pos {
		tiles[x] = make([][]*Tile, len(ys))
		for y, zs := range ys {
			var stack []*Tile
			for _, possible := range zs {
				if len(possible) == 1 {
					stack = append(stack, possible[0].Tile)
				}
			}
			tiles[x][y] = stack
		}
	}
	return tiles
}

func (g *Generator) Debug() string {
	buf := new(bytes.Buffer)
	for x, xs := range g.Pos {
		fmt.Fprintf(buf, "%2d:", x)
		for _, ys := range xs {
			fmt.Fprintf(buf, "%4d", len(ys))
		}
		buf.WriteRune('\n')
	}
	return buf.String()
}
