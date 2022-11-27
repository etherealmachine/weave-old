package main

// tileset name -> tile index -> Direction -> adjacent tiles
type AdjacencySet map[string]map[int]map[Direction][]*Tile

// x, y, z -> possible tiles in that position
type PossibilitySet [][][][]*Tile

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
		pos[x] = make([][][]*Tile, h)
		for y := 0; y < h; y++ {
			pos[x][y] = make([][]*Tile, d)
			pos[x][y][0] = firstLayer
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
			Tileset: "empty",
			Index:   0,
		}
	}
	if t2 == nil {
		t2 = &Tile{
			Tileset: "empty",
			Index:   0,
		}
	}
	if adj[t1.Tileset] == nil {
		adj[t1.Tileset] = make(map[int]map[Direction][]*Tile)
	}
	if adj[t1.Tileset][t1.Index] == nil {
		adj[t1.Tileset][t1.Index] = make(map[Direction][]*Tile)
	}
	for _, t := range adj[t1.Tileset][t1.Index][direction] {
		if t.Tileset == t2.Tileset && t.Index == t2.Index {
			return
		}
	}
	adj[t1.Tileset][t1.Index][direction] = append(adj[t1.Tileset][t1.Index][direction], t2)
	addEdge(adj, t2, direction.Inverse(), t1)
}

func (g *Generator) updatePossible(x, y, z int, dir Direction, offset [3]int, tile *Tile) bool {
	//adj := g.Adj[tile.Tileset][tile.Index][dir]
	//pos := g.Pos[x+offset[0]][y+offset[0]][z+offset[0]]
	return false
}

func (g *Generator) Generate() [][][]*Tile {
	for {
		changed := false
		for x, ys := range g.Pos {
			for y, zs := range ys {
				for z, possible := range zs {
					for _, tile := range possible {
						for dir, offset := range neighborOffsets {
							changed = changed || g.updatePossible(x, y, z, dir, offset, tile)
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
					stack = append(stack, possible[0])
				}
			}
			tiles[x][y] = stack
		}
	}
	return tiles
}
