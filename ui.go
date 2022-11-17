package main

import (
	"image"
	"log"
	"math"

	"github.com/etherealmachine/bento"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type UI struct {
	SelectedTileIndex int
	Tileset           *ebiten.Image
	Tiles             []*ebiten.Image
	Width             int
	SelectedLayer     int
	Layers            []*Layer
}

type Layer struct {
	Name  string
	Tiles []int
}

func (ui *UI) Draw(img *ebiten.Image) {
	if ui.Tiles != nil && ui.SelectedTileIndex > 0 {
		x, y := ebiten.CursorPosition()
		op := new(ebiten.DrawImageOptions)
		op.GeoM.Translate(math.Floor(float64(x)/16)*16, math.Floor(float64(y)/16)*16)
		img.DrawImage(ui.Tiles[ui.SelectedTileIndex-1], op)
	}
	if ui.SelectedLayer >= len(ui.Layers) {
		return
	}
	layer := ui.Layers[ui.SelectedLayer]
	for y := 0; y < len(layer.Tiles)/ui.Width; y++ {
		for x := 0; x < ui.Width; x++ {
			if layer.Tiles[y*ui.Width+x] > 0 {
				op := new(ebiten.DrawImageOptions)
				op.GeoM.Translate(float64(x*16), float64(y*16))
				img.DrawImage(ui.Tiles[layer.Tiles[y*ui.Width+x]-1], op)
			}
		}
	}
}

func (ui *UI) Update() bool {
	return false
}

func (ui *UI) AddLayer(event *bento.Event) {
	mapSize := len(ui.Layers[0].Tiles)
	ui.Layers = append(ui.Layers, &Layer{
		Name:  "Layer 1",
		Tiles: make([]int, mapSize),
	})
}

func (ui *UI) PasteTile(event *bento.Event) {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		return
	}
	if ui.SelectedTileIndex > 0 {
		tileX := event.X / 16
		tileY := event.Y / 16
		ui.Layers[ui.SelectedLayer].Tiles[tileY*ui.Width+tileX] = ui.SelectedTileIndex
	}
}

func (ui *UI) SelectTile(event *bento.Event) {
	if ui.Tileset == nil {
		ui.initializeTileset(event.Box.Bounds())
	}
	tileX := event.X / 17
	tileY := event.Y / 17
	ui.SelectedTileIndex = (tileY*ui.Width + tileX) + 1
}

func (ui *UI) initializeTileset(mapBounds image.Rectangle) {
	var err error
	ui.Tileset, _, err = ebitenutil.NewImageFromFile("dungeon.png")
	if err != nil {
		log.Fatal(err)
	}
	bounds := ui.Tileset.Bounds()
	width := bounds.Dx() / 17
	height := bounds.Dy() / 17
	ui.Width = width
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			tile := ebiten.NewImageFromImage(
				ui.Tileset.SubImage(
					image.Rect(x*17, y*17, (x+1)*17, (y+1)*17)))
			ui.Tiles = append(ui.Tiles, tile)
		}
	}
	ui.Layers = []*Layer{
		{
			Name:  "Layer 1",
			Tiles: make([]int, (mapBounds.Dx()/16)*(mapBounds.Dy()/16)),
		},
	}
}

func (ui *UI) UI() string {
	return `<col grow="1">
		<row grow="1">
			<col grow="1">
				<canvas grow="1" onDraw="Draw" onHover="PasteTile" />
			</col>
			<col grow="0 1">
				{{ range .Layers }}
					<button onClick="SelectLayer" btn="button.png 6" color="#ffffff" margin="4px" padding="12px">{{ .Name }}</button>
				{{ end }}
				<button onClick="AddLayer" color="#ffffff" margin="4px" padding="12px" btn="button.png 6">Add Layer</button>
			</col>
		</row>
		<row grow="1 0" justify="end" margin="16px">
			<img onClick="SelectTile" src="dungeon.png" />
		</row>
	</col>`
}
