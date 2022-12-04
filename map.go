package main

import (
	"encoding/json"
	"image"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Map struct {
	TileWidth, TileHeight int
	Spritesheets          map[string]*Spritesheet
	Tilemap               Tilemap
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
	subMap := make(map[int]map[int]Stack)
	for x := rect.Min.X; x < rect.Max.X; x++ {
		for y := rect.Min.Y; y < rect.Max.Y; y++ {
			stack := m.Tilemap[x][y]
			if stack != nil {
				if subMap[x-rect.Min.X] == nil {
					subMap[x-rect.Min.X] = make(map[int]Stack)
				}
				subMap[x-rect.Min.X][y-rect.Min.Y] = stack
			}
		}
	}
	//g := NewWFC(Analyze(m.Tilemap), rect.Dx(), rect.Dy(), subMap, seed)
	g := NewGreedyBFS(Analyze(m.Tilemap), rect.Dx(), rect.Dy(), subMap, seed)
	for !g.Done() {
	}
	result := g.Result()
	for x := 0; x < len(result); x++ {
		for y := 0; y < len(result[x]); y++ {
			if result[x][y] != nil {
				if len(m.Tilemap[x+rect.Min.X]) == 0 {
					m.Tilemap[x+rect.Min.X] = make(map[int]Stack)
				}
				m.Tilemap[x+rect.Min.X][y+rect.Min.Y] = result[x][y]
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
