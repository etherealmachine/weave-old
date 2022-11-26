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
	for x, ys := range m.Tiles {
		if len(ys) == 0 {
			delete(m.Tiles, x)
		}
		for y, tiles := range ys {
			if len(tiles) == 0 {
				delete(m.Tiles[x], y)
			}
		}
	}
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

func (m *Tilemap) SetTile(tileset string, index, x, y int, replace bool, z int) {
	if m.Tiles[x] == nil {
		m.Tiles[x] = make(map[int][]*Tile)
	}
	l := len(m.Tiles[x][y])
	if l == 0 {
		// first tile in the stack
		m.Tiles[x][y] = []*Tile{{
			Tileset: tileset,
			Index:   index,
		}}
	} else if z >= l {
		// append
		m.Tiles[x][y] = append(m.Tiles[x][y], &Tile{
			Tileset: tileset,
			Index:   index,
		})
	} else if replace {
		// replace
		m.Tiles[x][y][z] = &Tile{
			Tileset: tileset,
			Index:   index,
		}
	} else {
		// insert
		m.Tiles[x][y] = append(m.Tiles[x][y][:z+1], m.Tiles[x][y][z:]...)
		m.Tiles[x][y][z] = &Tile{
			Tileset: tileset,
			Index:   index,
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

func (s *Tileset) GetTile(index int) *ebiten.Image {
	if s == nil || index <= 0 {
		return nil
	}
	if s.tiles[index] == nil {
		rect := s.GetTileRect(index)
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

func (s *Tileset) GetTileRect(index int) *image.Rectangle {
	if s == nil || index <= 0 {
		return nil
	}
	w := s.Size + s.Spacing
	x := ((index - 1) % s.Width) * w
	y := ((index - 1) / s.Width) * w
	rect := image.Rect(x, y, x+s.Size, y+s.Size)
	return &rect
}
