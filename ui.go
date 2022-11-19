package main

import (
	"encoding/json"
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"sort"
	"strconv"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type UI struct {
	SelectedTileIndex int
	TileWidth         int
	Tileset           *ebiten.Image
	Tiles             []*ebiten.Image
	SelectedLayer     int
	Layers            []*Layer
	Adjacent          map[int][]map[int]bool
}

type Layer struct {
	Name   string
	Width  int
	Height int
	Tiles  []int
}

func (ui *UI) Draw(img *ebiten.Image) {
	if ui.Tiles == nil {
		ui.initialize(img.Bounds())
		ui.Load()
	}
	for i, layer := range ui.Layers {
		for y := 0; y < layer.Height; y++ {
			for x := 0; x < layer.Width; x++ {
				if layer.Tiles[y*layer.Width+x] > 0 {
					op := new(ebiten.DrawImageOptions)
					op.GeoM.Scale(2, 2)
					op.GeoM.Translate(float64(x*32), float64(y*32))
					tile := layer.Tiles[y*layer.Width+x] - 1
					img.DrawImage(ui.Tiles[tile], op)
					if ui.tileAt(x, y+1, i) == 0 {
						for _, adj := range ui.adj(tile, down) {
							op.GeoM.Translate(0, 32)
							op.ColorM.Scale(1, 1, 1, 0.5)
							img.DrawImage(ui.Tiles[adj], op)
							break
						}
					}
				}
			}
		}
	}
	if ui.Tiles != nil && ui.SelectedTileIndex > 0 {
		x, y := ebiten.CursorPosition()
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Scale(2, 2)
		op.GeoM.Translate(math.Floor(float64(x)/32)*32, math.Floor(float64(y)/32)*32)
		img.DrawImage(ui.Tiles[ui.SelectedTileIndex-1], op)
	}
}

func (ui *UI) Update(event *bento.Event) bool {
	_, sy := ebiten.Wheel()
	if sy != 0 {
		log.Println(sy)
	}
	return false
}

func (ui *UI) AddLayer(event *bento.Event) {
	mapSize := len(ui.Layers[0].Tiles)
	ui.Layers = append(ui.Layers, &Layer{
		Name:   fmt.Sprintf("Layer %d", len(ui.Layers)+1),
		Width:  ui.Layers[0].Width,
		Height: ui.Layers[0].Height,
		Tiles:  make([]int, mapSize),
	})
}

func (ui *UI) PasteTile(event *bento.Event) {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return
	}
	if ui.SelectedTileIndex > 0 {
		tileX := event.X / 32
		tileY := event.Y / 32
		layer := ui.Layers[ui.SelectedLayer]
		layer.Tiles[tileY*layer.Width+tileX] = ui.SelectedTileIndex
		ui.Save()
	}
}

func (ui *UI) SelectLayer(event *bento.Event) {
	i, _ := strconv.Atoi(event.Box.Attrs["index"])
	ui.SelectedLayer = i
}

func (ui *UI) SelectTile(event *bento.Event) {
	tileX := event.X / 34
	tileY := event.Y / 34
	ui.SelectedTileIndex = (tileY*ui.TileWidth + tileX) + 1
}

func (ui *UI) initialize(mapBounds image.Rectangle) {
	var err error
	ui.Tileset, _, err = ebitenutil.NewImageFromFile("dungeon.png")
	if err != nil {
		log.Fatal(err)
	}
	bounds := ui.Tileset.Bounds()
	width := bounds.Dx() / 17
	ui.TileWidth = width
	height := bounds.Dy() / 17
	for y := 0; y <= height; y++ {
		for x := 0; x < width; x++ {
			tile := ebiten.NewImageFromImage(
				ui.Tileset.SubImage(
					image.Rect(x*17, y*17, (x+1)*17, (y+1)*17)))
			ui.Tiles = append(ui.Tiles, tile)
		}
	}
	mapWidth := mapBounds.Dx() / 16
	mapHeight := mapBounds.Dy() / 16
	ui.Layers = []*Layer{
		{
			Name:   "Layer 1",
			Width:  mapWidth,
			Height: mapHeight,
			Tiles:  make([]int, mapWidth*mapHeight),
		},
	}
	ui.Adjacent = make(map[int][]map[int]bool)
}

func (ui *UI) Save() {
	f, err := os.Create("map.json")
	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(ui.Layers); err != nil {
		log.Fatal(err)
	}
	ui.analyze()
}

func (ui *UI) Load() {
	f, err := os.Open("map.json")
	if err != nil {
		return
	}
	d := json.NewDecoder(f)
	if err := d.Decode(&ui.Layers); err != nil {
		log.Fatal(err)
	}
	ui.analyze()
}

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

func (ui *UI) UI() string {
	return `<col grow="1">
		<row grow="1">
			<col grow="1">
				<canvas grow="1" onDraw="Draw" onHover="PasteTile" onUpdate="Update" />
			</col>
			<col grow="0 1">
				{{ range $i, $layer := .Layers }}
					<button
							onClick="SelectLayer"
							index="{{ $i }}"
							btn="button.png 6"
							color="#ffffff"
							margin="4px"
							padding="12px"
							underline="{{ eq $i $.SelectedLayer }}">{{ $layer.Name }}</button>
				{{ end }}
				<button onClick="AddLayer" color="#ffffff" margin="4px" padding="12px" btn="button.png 6">Add Layer</button>
			</col>
		</row>
		<row grow="1 0" justify="end" margin="16px">
			<img onClick="SelectTile" src="dungeon.png" scale="2" />
		</row>
	</col>`
}
