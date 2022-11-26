package main

import (
	"encoding/json"
	"image"
	"log"
	"math"
	"math/rand"
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
	m.Cleanup()
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

func (m *Tilemap) Erase(rect image.Rectangle) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			if m.Tiles[x] != nil {
				m.Tiles[x][y] = nil
			}
		}
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
				if tile.Index <= 0 || m.Tilesets[tile.Tileset] == nil {
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

type Graph map[string]map[int]map[Direction][]*Tile

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

func (g Graph) AddEdge(t1 *Tile, direction Direction, t2 *Tile) {
	if t1 == nil || t2 == nil {
		return
	}
	if g[t1.Tileset] == nil {
		g[t1.Tileset] = make(map[int]map[Direction][]*Tile)
	}
	if g[t1.Tileset][t1.Index] == nil {
		g[t1.Tileset][t1.Index] = make(map[Direction][]*Tile)
	}
	for _, t := range g[t1.Tileset][t1.Index][direction] {
		if t.Tileset == t2.Tileset && t.Index == t2.Index {
			return
		}
	}
	g[t1.Tileset][t1.Index][direction] = append(g[t1.Tileset][t1.Index][direction], t2)
	g.AddEdge(t2, direction.Inverse(), t1)
}

func (g Graph) AllTiles() []*Tile {
	var tiles []*Tile
	for _, tileset := range g {
		for _, adj := range tileset {
			for _, ts := range adj {
				for _, t := range ts {
					exists := false
					for _, v := range tiles {
						if t.Tileset == v.Tileset && t.Index != v.Index {
							exists = true
							break
						}
					}
					if !exists {
						tiles = append(tiles, t)
					}
				}
			}
		}
	}
	return tiles
}

func (m *Tilemap) Analyze() {
	m.Adjacencies = make(map[string]map[int]map[Direction][]*Tile)
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

func (m *Tilemap) Generate(rect image.Rectangle) {
	w, h := rect.Dx(), rect.Dy()
	possibilities := make([][]*Tile, w*h)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			possibilities[y*w+x] = m.Adjacencies.AllTiles()
		}
	}
	for {
		minCount, minIndex := math.MaxInt, -1
		for x := 0; x < w; x++ {
			for y := 0; y < h; y++ {
				count := len(possibilities[y*w+x])
				if count > 1 && count < minCount {
					minCount = count
					minIndex = y*w + x
				}
			}
		}
		if minIndex == -1 {
			break
		}
		rand.Shuffle(len(possibilities[minIndex]), func(i, j int) {
			possibilities[minIndex][i], possibilities[minIndex][j] = possibilities[minIndex][j], possibilities[minIndex][i]
		})
		possibilities[minIndex] = possibilities[minIndex][0:1]
	}
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			if m.Tiles[x+rect.Min.X] == nil {
				m.Tiles[x+rect.Min.X] = make(map[int][]*Tile)
			}
			m.Tiles[x+rect.Min.X][y+rect.Min.Y] = possibilities[y*w+x]
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
