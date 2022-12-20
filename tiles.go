package main

import (
	"fmt"
	"image"
	"log"
	"path"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Tileset struct {
	Spritesheets map[string]*Spritesheet
	tiles        []*Tile
}

func NewTileset(prefix string) *Tileset {
	ts := &Tileset{
		Spritesheets: make(map[string]*Spritesheet),
	}
	if err := ts.Add(path.Join(prefix, "dungeon.png"), 16, 1); err != nil {
		log.Fatal(err)
	}
	if err := ts.Add(path.Join(prefix, "general.png"), 16, 1); err != nil {
		log.Fatal(err)
	}
	if err := ts.Add(path.Join(prefix, "indoors.png"), 16, 1); err != nil {
		log.Fatal(err)
	}
	if err := ts.Add(path.Join(prefix, "characters.png"), 16, 1); err != nil {
		log.Fatal(err)
	}
	return ts
}

func (ts *Tileset) Add(filename string, size, spacing int) error {
	var err error
	ts.Spritesheets[filename], err = NewSpritesheet(filename, size, spacing)
	ts.tiles = nil
	return err
}

func (ts *Tileset) Image(t *Tile) *ebiten.Image {
	if t == nil {
		return nil
	}
	if ts.Spritesheets[t.Spritesheet] == nil {
		return nil
	}
	return ts.Spritesheets[t.Spritesheet].Image(t.Index)
}

func (ts *Tileset) Tiles() []*Tile {
	if ts.tiles != nil {
		return ts.tiles
	}
	var tiles []*Tile
	for name, sheet := range ts.Spritesheets {
		for i, img := range sheet.Tiles() {
			tiles = append(tiles, &Tile{
				Spritesheet: name,
				Index:       i,
				Image:       img,
			})
		}
	}
	ts.tiles = tiles
	return tiles
}

type Spritesheet struct {
	Name          string
	Img           *ebiten.Image
	Size          int
	Spacing       int
	Width, Height int
	tiles         []*ebiten.Image
}

func NewSpritesheet(filename string, size, spacing int) (*Spritesheet, error) {
	img, _, err := ebitenutil.NewImageFromFile(filename)
	if err != nil {
		return nil, err
	}
	w := size + spacing
	bounds := img.Bounds()
	width, height := (bounds.Dx()/w)+1, (bounds.Dy()/w)+1
	return &Spritesheet{
		Name:    filename,
		Img:     img,
		Size:    size,
		Spacing: spacing,
		Width:   width,
		Height:  height,
		tiles:   make([]*ebiten.Image, width*height),
	}, nil
}

func (s *Spritesheet) Image(index int) *ebiten.Image {
	if s == nil || index < 0 || index >= len(s.tiles) {
		return nil
	}
	if s.tiles[index] == nil {
		rect := s.Rect(index)
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

func (s *Spritesheet) Rect(index int) *image.Rectangle {
	if s == nil || index < 0 {
		return nil
	}
	w := s.Size + s.Spacing
	x := (index % s.Width) * w
	y := (index / s.Width) * w
	rect := image.Rect(x, y, x+s.Size, y+s.Size)
	return &rect
}

func (s *Spritesheet) Tiles() []*ebiten.Image {
	for i, img := range s.tiles {
		if img == nil {
			s.tiles[i] = s.Image(i)
		}
	}
	return s.tiles
}

type Tile struct {
	Spritesheet string
	Index       int
	Image       *ebiten.Image
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
