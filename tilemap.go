package main

import (
	"encoding/json"
	"image"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Tilemap struct {
	TileWidth, TileHeight int
	Tilesets              map[string]*Tileset
	Tiles                 map[int]map[int][]*Tile
	Adjacencies           Graph
}

type Tile struct {
	Tileset string
	Index   int
}

func NewTilemap(w, h int) *Tilemap {
	t := &Tilemap{
		TileWidth:  w,
		TileHeight: h,
		Tilesets:   make(map[string]*Tileset),
		Tiles:      make(map[int]map[int][]*Tile),
	}
	if err := t.Load("map.json"); err != nil {
		log.Fatal(err)
	}
	return t
}

func (m *Tilemap) AddTileset(filename string, size, spacing int) error {
	var err error
	m.Tilesets[filename], err = NewTileset(filename, size, spacing)
	return err
}

func (m *Tilemap) Save(filename string) error {
	m.Cleanup()
	m.Analyze()
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(m); err != nil {
		return err
	}
	return nil
}

func (m *Tilemap) Load(filename string) error {
	f, err := os.Open(filename)
	if os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return err
	}
	d := json.NewDecoder(f)
	if err := d.Decode(m); err != nil {
		return err
	}
	m.Analyze()
	return nil
}

func (m *Tilemap) SetTile(tile *Tile, x, y int, replace bool, z int) {
	if m.Tiles[x] == nil {
		m.Tiles[x] = make(map[int][]*Tile)
	}
	l := len(m.Tiles[x][y])
	if l == 0 {
		// first tile in the stack
		m.Tiles[x][y] = []*Tile{tile}
	} else if z >= l {
		// append
		m.Tiles[x][y] = append(m.Tiles[x][y], tile)
	} else if replace {
		// replace
		m.Tiles[x][y][z] = tile
	} else {
		// insert
		m.Tiles[x][y] = append(m.Tiles[x][y][:z+1], m.Tiles[x][y][z:]...)
		m.Tiles[x][y][z] = tile
	}
	if err := m.Save("map.json"); err != nil {
		log.Fatal(err)
	}
}

func (m *Tilemap) EraseTile(x, y int) {
	if l := len(m.Tiles[x][y]); l > 0 {
		m.Tiles[x][y] = m.Tiles[x][y][:l-1]
	}
	if err := m.Save("map.json"); err != nil {
		log.Fatal(err)
	}
}

func (m *Tilemap) TileImage(t *Tile) *ebiten.Image {
	if t == nil {
		return nil
	}
	if m.Tilesets[t.Tileset] == nil {
		return nil
	}
	return m.Tilesets[t.Tileset].TileImage(t.Index)
}

func (m *Tilemap) TileAt(x, y, z int) *Tile {
	if len(m.Tiles[x]) == 0 {
		return nil
	}
	if len(m.Tiles[x][y]) == 0 {
		return nil
	}
	if z < 0 || z >= len(m.Tiles[x][y]) {
		return nil
	}
	return m.Tiles[x][y][z]
}

func (m *Tilemap) Cleanup() {
	for x, ys := range m.Tiles {
		for y, tiles := range ys {
			for z, tile := range tiles {
				if m.Tilesets[tile.Tileset] == nil {
					m.Tiles[x][y] = append(m.Tiles[x][y][:z], m.Tiles[x][y][z+1:]...)
				}
			}
			if len(tiles) == 0 {
				delete(m.Tiles[x], y)
			}
		}
		if len(ys) == 0 {
			delete(m.Tiles, x)
		}
	}
}

type Graph map[string]map[int]map[string][]*Tile

var neighborOffsets = map[string][3]int{
	"above": {0, 0, 1},
	"below": {0, 0, -1},
	"north": {0, 1, 0},
	"south": {0, -1, 0},
	"east":  {1, 0, 0},
	"west":  {-1, 0, 0},
}

func (g Graph) AddEdge(t1 *Tile, direction string, t2 *Tile) {
	if t1 == nil || t2 == nil {
		return
	}
	if g[t1.Tileset] == nil {
		g[t1.Tileset] = make(map[int]map[string][]*Tile)
	}
	if g[t1.Tileset][t1.Index] == nil {
		g[t1.Tileset][t1.Index] = make(map[string][]*Tile)
	}
	for _, t := range g[t1.Tileset][t1.Index][direction] {
		if t.Tileset == t2.Tileset && t.Index == t2.Index {
			return
		}
	}
	g[t1.Tileset][t1.Index][direction] = append(g[t1.Tileset][t1.Index][direction], t2)
}

func (m *Tilemap) Analyze() {
	m.Adjacencies = make(map[string]map[int]map[string][]*Tile)
	for x, ys := range m.Tiles {
		for y, tiles := range ys {
			for z, tile := range tiles {
				for dir, offset := range neighborOffsets {
					m.Adjacencies.AddEdge(m.TileAt(x+offset[0], y+offset[1], z+offset[2]), dir, tile)
				}
			}
		}
	}
}

type Tileset struct {
	Name          string
	Img           *ebiten.Image
	Size          int
	Spacing       int
	Width, Height int
	tiles         map[int]*ebiten.Image
}

func NewTileset(filename string, size, spacing int) (*Tileset, error) {
	img, _, err := ebitenutil.NewImageFromFile(filename)
	if err != nil {
		return nil, err
	}
	w := size + spacing
	bounds := img.Bounds()
	return &Tileset{
		Name:    filename,
		Img:     img,
		Size:    size,
		Spacing: spacing,
		Width:   (bounds.Dx() / w) + 1,
		Height:  (bounds.Dy() / w) + 1,
		tiles:   make(map[int]*ebiten.Image),
	}, nil
}

func (s *Tileset) TileImage(index int) *ebiten.Image {
	if s == nil || index <= 0 {
		return nil
	}
	if s.tiles[index] == nil {
		rect := s.TileRect(index)
		s.tiles[index] = ebiten.NewImageFromImage(s.Img.SubImage(*rect))
	}
	return s.tiles[index]
}

func (s *Tileset) TileAt(x, y int) int {
	if s == nil {
		return 0
	}
	w := s.Size + s.Spacing
	return (y/w)*s.Width + (x / w) + 1
}

func (s *Tileset) TileRect(index int) *image.Rectangle {
	if s == nil || index <= 0 {
		return nil
	}
	w := s.Size + s.Spacing
	x := ((index - 1) % s.Width) * w
	y := ((index - 1) / s.Width) * w
	rect := image.Rect(x, y, x+s.Size, y+s.Size)
	return &rect
}
