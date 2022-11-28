package main

type Tileset map[string]map[int]bool

func NewTileset(tiles ...*Tile) Tileset {
	ts := make(Tileset)
	for _, t := range tiles {
		ts.Add(t)
	}
	return ts
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

// x, y, z -> possible tiles in that position
type PossibilitySet [][][]Tileset

func (s PossibilitySet) At(x, y, z int) Tileset {
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

type Generator struct {
	Adj AdjacencySet
	Pos PossibilitySet
}

func NewGenerator(m *Tilemap, x, y, w, h int) *Generator {
	adj := make(AdjacencySet)
	d := 0
	var firstLayer []*Tile
	for x, ys := range m.Tiles {
		for y, tiles := range ys {
			for z, tile := range tiles {
				if z > d {
					d = z
				}
				if z == 0 {
					firstLayer = append(firstLayer, tile)
				}
				for dir, offset := range neighborOffsets {
					addEdge(adj, m.TileAt(x+offset[0], y+offset[1], z+offset[2]), dir, tile)
				}
			}
		}
	}
	pos := make(PossibilitySet, w)
	for x := 0; x < w; x++ {
		pos[x] = make([][]Tileset, h)
		for y := 0; y < h; y++ {
			pos[x][y] = make([]Tileset, d)
			pos[x][y][0] = NewTileset(firstLayer...)
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

func (g *Generator) Generate() [][][]*Tile {
	for {
		changed := false
		for x, ys := range g.Pos {
			for y, zs := range ys {
				for z, possible := range zs {
					for _, tile := range possible.Tiles() {
						for dir, offset := range neighborOffsets {
							nx, ny, nz := x+offset[0], y+offset[1], z+offset[2]
							adj := g.Adj.At(tile.Spritesheet, tile.Index, dir)
							pos := g.Pos.At(nx, ny, nz)
							if pos != nil {
								changed = changed || pos.Union(adj)
							}
						}
					}
				}
			}
		}
		if !changed {
			break
		}
	}
	tiles := make([][][]*Tile, len(g.Pos))
	for x, ys := range g.Pos {
		tiles[x] = make([][]*Tile, len(ys))
		for y, zs := range ys {
			var stack []*Tile
			for _, possible := range zs {
				if len(possible) == 1 {
					stack = append(stack, possible.Tiles()[0])
				}
			}
			tiles[x][y] = stack
		}
	}
	return tiles
}
