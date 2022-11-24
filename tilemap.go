package main

import (
	"encoding/json"
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Tilemap struct {
	Tilesets map[string]*Tileset
}

func NewTilemap() *Tilemap {
	return &Tilemap{
		Tilesets: make(map[string]*Tileset),
	}
}

func (m *Tilemap) AddTileset(filename string, size, spacing int) error {
	var err error
	m.Tilesets[filename], err = NewTileset(filename, size, spacing)
	return err
}

func (m *Tilemap) Save(filename string) error {
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
	return nil
}

func (s *Tileset) GetTile(index int) *ebiten.Image {
	if s == nil || index <= 0 {
		return nil
	}
	if s.tiles[index] == nil {
		w := s.Size + s.Spacing
		x := ((index - 1) % s.Width) * w
		y := ((index - 1) / s.Width) * w
		s.tiles[index] = ebiten.NewImageFromImage(s.Img.SubImage(image.Rect(x, y, x+s.Size, y+s.Size)))
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

/*
func (ui *UI) tileAt(x, y, i int) int {
	if i < 0 || i >= len(ui.Layers) {
		return 0
	}
	layer := ui.Layers[i]
	if x < 0 || x >= layer.Width {
		return 0
	}
	if y < 0 || y >= layer.Height {
		return 0
	}
	return layer.Tiles[y*layer.Width+x]
}

type direction int

const (
	below = direction(0)
	above = direction(1)
	up    = direction(2)
	down  = direction(3)
	left  = direction(4)
	right = direction(5)
)

func (ui *UI) addEdge(t1 int, dir direction, t2 int) {
	if t1 == 0 || t2 == 0 {
		return
	}
	if dirs := ui.Adjacent[t1]; dirs == nil {
		ui.Adjacent[t1] = make([]map[int]bool, 6)
		for i := 0; i < 6; i++ {
			ui.Adjacent[t1][i] = make(map[int]bool)
		}
	}
	ui.Adjacent[t1][dir][t2] = true
}

func (ui *UI) adj(t1 int, dir direction) []int {
	if ui.Adjacent[t1] == nil {
		return nil
	}
	if dir < 0 || int(dir) >= len(ui.Adjacent[t1]) {
		return nil
	}
	tiles := make([]int, len(ui.Adjacent[t1][dir]))
	i := 0
	for t := range ui.Adjacent[t1][dir] {
		tiles[i] = t
		i++
	}
	sort.Ints(tiles)
	return tiles
}

func (ui *UI) analyze() {
	for i, layer := range ui.Layers {
		for y := 0; y < layer.Height; y++ {
			for x := 0; x < layer.Width; x++ {
				if tile := layer.Tiles[y*layer.Width+x] - 1; tile > 0 {
					ui.addEdge(tile, below, ui.tileAt(x, y, i-1))
					ui.addEdge(tile, above, ui.tileAt(x, y, i+1))
					ui.addEdge(tile, up, ui.tileAt(x, y-1, i))
					ui.addEdge(tile, down, ui.tileAt(x, y+1, i))
					ui.addEdge(tile, left, ui.tileAt(x-1, y, i))
					ui.addEdge(tile, right, ui.tileAt(x+1, y, i))
				}
			}
		}
	}
}
*/
