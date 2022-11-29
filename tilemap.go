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
	Spritesheets          map[string]*Spritesheet
	Tiles                 map[int]map[int][]*Tile
}

type Tile struct {
	Spritesheet string
	Index       int
}

func NewTilemap(w, h int) *Tilemap {
	t := &Tilemap{
		TileWidth:    w,
		TileHeight:   h,
		Spritesheets: make(map[string]*Spritesheet),
		Tiles:        make(map[int]map[int][]*Tile),
	}
	if err := t.Load("map.json"); err != nil {
		log.Fatal(err)
	}
	return t
}

func (m *Tilemap) AddTileset(filename string, size, spacing int) error {
	var err error
	m.Spritesheets[filename], err = NewSpritesheet(filename, size, spacing)
	return err
}

func (m *Tilemap) Save(filename string) error {
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
	if m.Spritesheets[t.Spritesheet] == nil {
		return nil
	}
	return m.Spritesheets[t.Spritesheet].TileImage(t.Index)
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
			var stack []*Tile
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
			m.Tiles[x][y] = stack
			if len(stack) == 0 {
				delete(m.Tiles[x], y)
			}
		}
		if len(ys) == 0 {
			delete(m.Tiles, x)
		}
	}
}

func (m *Tilemap) Generate(rect image.Rectangle) {
	g := NewGenerator(m, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
	tiles := g.Generate()
	for x := 0; x < len(tiles); x++ {
		for y := 0; y < len(tiles[x]); y++ {
			if len(m.Tiles[x+rect.Min.X]) == 0 {
				m.Tiles[x+rect.Min.X] = make(map[int][]*Tile)
			}
			m.Tiles[x+rect.Min.X][y+rect.Min.Y] = tiles[x][y]
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
