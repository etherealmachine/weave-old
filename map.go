package main

import (
	"encoding/json"
	"image"
	"log"
	"os"
)

type Map struct {
	*Tileset
	TileWidth, TileHeight int
	Tilemap               Tilemap
}

func NewMap(w, h int, tileset *Tileset) *Map {
	t := &Map{
		TileWidth:  w,
		TileHeight: h,
		Tileset:    tileset,
		Tilemap:    make(map[int]map[int]Stack),
	}
	/*
		if err := t.Load("map.json"); err != nil {
			log.Fatal(err)
		}
	*/
	return t
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
