package main

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Map struct {
	TileWidth, TileHeight int
	Spritesheets          map[string]*Spritesheet
	Tilemap               Tilemap
}

type Tilemap map[int]map[int]Stack

func (m Tilemap) Set(tile *Tile, x, y int, replace bool, z int) {
	if m[x] == nil {
		m[x] = make(map[int]Stack)
	}
	l := len(m[x][y])
	if l == 0 {
		// first tile in the stack
		m[x][y] = Stack{tile}
	} else if z >= l {
		// append
		m[x][y] = append(m[x][y], tile)
	} else if replace {
		// replace
		m[x][y][z] = tile
	} else {
		// insert
		m[x][y] = append(m[x][y][:z+1], m[x][y][z:]...)
		m[x][y][z] = tile
	}
}

func (m Tilemap) At(x, y, z int) *Tile {
	if len(m[x]) == 0 {
		return nil
	}
	if len(m[x][y]) == 0 {
		return nil
	}
	if z < 0 || z >= len(m[x][y]) {
		return nil
	}
	return m[x][y][z]
}

type Tile struct {
	Spritesheet string
	Index       int
}

func (t *Tile) Hash() string {
	return fmt.Sprintf("%s:%d", t.Spritesheet, t.Index)
}

type Stack []*Tile

func (s Stack) Hash() string {
	a := make([]string, len(s))
	for i, t := range s {
		a[i] = t.Hash()
	}
	return strings.Join(a, ",")
}

func NewMap(w, h int) *Map {
	t := &Map{
		TileWidth:    w,
		TileHeight:   h,
		Spritesheets: make(map[string]*Spritesheet),
		Tilemap:      make(map[int]map[int]Stack),
	}
	if err := t.Load("map.json"); err != nil {
		log.Fatal(err)
	}
	return t
}

func (m *Map) AddTileset(filename string, size, spacing int) error {
	var err error
	m.Spritesheets[filename], err = NewSpritesheet(filename, size, spacing)
	return err
}

func (m *Map) Save(filename string) error {
	m.Cleanup()
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

func (m *Map) Load(filename string) error {
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
	return nil
}

func (m *Map) SetTile() {

}

func (m *Map) Erase(rect image.Rectangle) {
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			if m.Tilemap[x] != nil {
				m.Tilemap[x][y] = nil
			}
		}
	}
	if err := m.Save("map.json"); err != nil {
		log.Fatal(err)
	}
}

func (m *Map) EraseTile(x, y int) {
	if l := len(m.Tilemap[x][y]); l > 0 {
		m.Tilemap[x][y] = m.Tilemap[x][y][:l-1]
	}
	if err := m.Save("map.json"); err != nil {
		log.Fatal(err)
	}
}

func (m *Map) TileImage(t *Tile) *ebiten.Image {
	if t == nil {
		return nil
	}
	if m.Spritesheets[t.Spritesheet] == nil {
		return nil
	}
	return m.Spritesheets[t.Spritesheet].TileImage(t.Index)
}

func (m *Map) Cleanup() {
	for x, ys := range m.Tilemap {
		for y, tiles := range ys {
			var stack Stack
			for _, tile := range tiles {
				if tile == nil || tile.Index <= 0 || m.Spritesheets[tile.Spritesheet] == nil {
					continue
				}
				if len(stack) > 0 {
					prev := stack[len(stack)-1]
					if prev.Spritesheet == tile.Spritesheet && prev.Index == tile.Index {
						continue
					}
				}
				stack = append(stack, tile)
			}
			m.Tilemap[x][y] = stack
			if len(stack) == 0 {
				delete(m.Tilemap[x], y)
			}
		}
		if len(ys) == 0 {
			delete(m.Tilemap, x)
		}
	}
}

func (m *Map) Generate(rect image.Rectangle, seed int64) {
	g := NewGenerator(m.Tilemap)
	g.Init(rect.Dx(), rect.Dy(), nil, seed)
	for !g.Done() {
	}
	for x := 0; x < g.Width; x++ {
		for y := 0; y < g.Height; y++ {
			if len(m.Tilemap[x+rect.Min.X]) == 0 {
				m.Tilemap[x+rect.Min.X] = make(map[int]Stack)
			}
			if i := g.Map.At(x, y); i != nil {
				m.Tilemap[x+rect.Min.X][y+rect.Min.Y] = g.Domain[*i]
			}
		}
	}
}

type Spritesheet struct {
	Name          string
	Img           *ebiten.Image
	Size          int
	Spacing       int
	Width, Height int
	tiles         map[int]*ebiten.Image
}

func NewSpritesheet(filename string, size, spacing int) (*Spritesheet, error) {
	img, _, err := ebitenutil.NewImageFromFile(filename)
	if err != nil {
		return nil, err
	}
	w := size + spacing
	bounds := img.Bounds()
	return &Spritesheet{
		Name:    filename,
		Img:     img,
		Size:    size,
		Spacing: spacing,
		Width:   (bounds.Dx() / w) + 1,
		Height:  (bounds.Dy() / w) + 1,
		tiles:   make(map[int]*ebiten.Image),
	}, nil
}

func (s *Spritesheet) TileImage(index int) *ebiten.Image {
	if s == nil || index <= 0 {
		return nil
	}
	if s.tiles[index] == nil {
		rect := s.TileRect(index)
		s.tiles[index] = ebiten.NewImageFromImage(s.Img.SubImage(*rect))
	}
	return s.tiles[index]
}

func (s *Spritesheet) TileAt(x, y int) int {
	if s == nil {
		return 0
	}
	w := s.Size + s.Spacing
	return (y/w)*s.Width + (x / w) + 1
}

func (s *Spritesheet) TileRect(index int) *image.Rectangle {
	if s == nil || index <= 0 {
		return nil
	}
	w := s.Size + s.Spacing
	x := ((index - 1) % s.Width) * w
	y := ((index - 1) / s.Width) * w
	rect := image.Rect(x, y, x+s.Size, y+s.Size)
	return &rect
}
